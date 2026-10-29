package main

// End-to-end test of Google/Apple sign-in against a fake identity provider and
// a real MongoDB. It is skipped unless MONGODB_TEST_URI is set, e.g.
//
//	docker run -d -p 127.0.0.1:27017:27017 mongo:7
//	MONGODB_TEST_URI=mongodb://127.0.0.1:27017 go test -run Social .

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"lldprep/internal/accounts"
	"lldprep/internal/content"
	"lldprep/internal/oauth"
)

type fakeProvider struct {
	t      *testing.T
	key    *rsa.PrivateKey
	srv    *httptest.Server
	mu     sync.Mutex
	nonces map[string]string // code -> nonce
	next   int
	// who the next person to "consent" is
	sub, email string
}

func newFakeProvider(t *testing.T) *fakeProvider {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeProvider{t: t, key: key, nonces: map[string]string{}}
	mux := http.NewServeMux()
	// The consent screen: approve immediately and send the browser back.
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f.mu.Lock()
		f.next++
		code := fmt.Sprintf("code-%d", f.next)
		f.nonces[code] = q.Get("nonce") + "|" + f.sub + "|" + f.email
		f.mu.Unlock()
		http.Redirect(w, r, q.Get("redirect_uri")+"?"+url.Values{"code": {code}, "state": {q.Get("state")}}.Encode(), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		f.mu.Lock()
		v, ok := f.nonces[r.PostForm.Get("code")]
		delete(f.nonces, r.PostForm.Get("code")) // codes are single-use
		f.mu.Unlock()
		if !ok {
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
			return
		}
		parts := strings.SplitN(v, "|", 3)
		json.NewEncoder(w).Encode(map[string]string{"id_token": f.sign(map[string]any{
			"iss": "https://idp.test", "aud": "client-1", "sub": parts[1], "email": parts[2],
			"email_verified": true, "nonce": parts[0], "exp": time.Now().Add(time.Hour).Unix(),
		})})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "kid": "k1",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeProvider) sign(claims map[string]any) string {
	h, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
	c, _ := json.Marshal(claims)
	in := base64.RawURLEncoding.EncodeToString(h) + "." + base64.RawURLEncoding.EncodeToString(c)
	d := sha256.Sum256([]byte(in))
	sig, err := rsa.SignPKCS1v15(rand.Reader, f.key, crypto.SHA256, d[:])
	if err != nil {
		f.t.Fatal(err)
	}
	return in + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (f *fakeProvider) provider(name string) *oauth.Provider {
	return &oauth.Provider{
		Name: name, ClientID: "client-1",
		AuthURL: f.srv.URL + "/auth", TokenURL: f.srv.URL + "/token", JWKSURL: f.srv.URL + "/keys",
		Issuers: []string{"https://idp.test"}, Scopes: []string{"openid", "email"}, UsePKCE: name == "google",
		ClientSecret: func() (string, error) { return "secret", nil },
	}
}

type testApp struct {
	t   *testing.T
	srv *httptest.Server
	idp *fakeProvider
}

func newTestApp(t *testing.T, invite string) *testApp {
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Skip("set MONGODB_TEST_URI to run the social sign-in integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbName := fmt.Sprintf("lldprep_test_%d", time.Now().UnixNano())
	db, err := accounts.Connect(ctx, uri, dbName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.DropForTests(context.Background()); db.Close() })
	idp := newFakeProvider(t)
	au := &auth{db: db, inviteCode: invite, limiter: newLoginLimiter(), social: &social{
		providers: map[string]*oauth.Provider{"google": idp.provider("google"), "apple": idp.provider("apple")},
		key:       []byte("test-key"),
	}}
	mux := http.NewServeMux()
	au.routes(mux)
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(userOf(r))) })
	srv := httptest.NewServer(au.requireUser(mux))
	t.Cleanup(srv.Close)
	return &testApp{t: t, srv: srv, idp: idp}
}

// browser is a cookie-keeping client that doesn't follow redirects by itself.
type browser struct {
	t *testing.T
	c *http.Client
}

func (a *testApp) browser() *browser {
	jar, _ := cookiejar.New(nil)
	return &browser{t: a.t, c: &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
}

// signIn runs start → provider consent → callback and returns where the app sent the browser.
func (b *browser) signIn(base, provider, query string) string {
	loc := base + "/api/auth/" + provider + "/start" + query
	for i := 0; i < 5; i++ {
		resp, err := b.c.Get(loc)
		if err != nil {
			b.t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			b.t.Fatalf("GET %s: %s", loc, resp.Status)
		}
		next := resp.Header.Get("Location")
		if !strings.HasPrefix(next, "http") { // back in the app
			return next
		}
		loc = next
	}
	b.t.Fatal("too many redirects")
	return ""
}

func (b *browser) me(base string) map[string]any {
	resp, err := b.c.Get(base + "/api/auth/me")
	if err != nil {
		b.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return out
}

func (b *browser) post(base, path, body string) int {
	resp, err := b.c.Post(base+path, "application/json", strings.NewReader(body))
	if err != nil {
		b.t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestSocialSignUpAndSignIn(t *testing.T) {
	a := newTestApp(t, "")
	a.idp.sub, a.idp.email = "g-1", "Asha.K@example.com"

	b := a.browser()
	if to := b.signIn(a.srv.URL, "google", ""); to != "/" {
		t.Fatalf("first sign-in redirected to %q", to)
	}
	me := b.me(a.srv.URL)
	if me["user"] != "asha.k" || fmt.Sprint(me["linked"]) != "[google]" || me["hasPassword"] != false {
		t.Fatalf("me after Google sign-up: %v", me)
	}
	if fmt.Sprint(me["providers"]) != "[apple google]" {
		t.Fatalf("providers: %v", me["providers"])
	}

	// A second browser signing in with the same Google account lands on the same user.
	b2 := a.browser()
	b2.signIn(a.srv.URL, "google", "")
	if u := b2.me(a.srv.URL)["user"]; u != "asha.k" {
		t.Fatalf("returning sign-in got user %v", u)
	}

	// A different person with a clashing email local part gets a suffixed username.
	a.idp.sub, a.idp.email = "g-2", "asha.k@other.com"
	b3 := a.browser()
	b3.signIn(a.srv.URL, "google", "")
	if u := b3.me(a.srv.URL)["user"]; u != "asha.k2" {
		t.Fatalf("second person got user %v", u)
	}

	// Password sign-in doesn't work for a Google-only account.
	if code := a.browser().post(a.srv.URL, "/api/auth/login", `{"username":"asha.k","password":"anything-at-all"}`); code != http.StatusUnauthorized {
		t.Fatalf("password login for external user: %d", code)
	}
}

func TestSocialRespectsInviteCode(t *testing.T) {
	a := newTestApp(t, "letmein")
	a.idp.sub, a.idp.email = "g-9", "kim@example.com"

	to := a.browser().signIn(a.srv.URL, "google", "")
	if !strings.Contains(to, "auth_error=") || !strings.Contains(to, "invite") {
		t.Fatalf("sign-up without invite should be refused, went to %q", to)
	}
	resp, _ := a.browser().c.Get(a.srv.URL + "/api/auth/google/start?invite=wrong")
	if loc := resp.Header.Get("Location"); !strings.Contains(loc, "Wrong+invite") {
		t.Fatalf("wrong invite went to %q", loc)
	}
	b := a.browser()
	if to := b.signIn(a.srv.URL, "google", "?invite=letmein"); to != "/" {
		t.Fatalf("sign-up with invite went to %q", to)
	}
	if u := b.me(a.srv.URL)["user"]; u != "kim" {
		t.Fatalf("user %v", u)
	}
	// Existing users don't need the invite again.
	b2 := a.browser()
	if to := b2.signIn(a.srv.URL, "google", ""); to != "/" || b2.me(a.srv.URL)["user"] != "kim" {
		t.Fatalf("returning user without invite went to %q", to)
	}
}

func TestSocialLinkToExistingAccount(t *testing.T) {
	a := newTestApp(t, "")
	b := a.browser()
	if code := b.post(a.srv.URL, "/api/auth/signup", `{"username":"ravi","password":"password123"}`); code != http.StatusOK {
		t.Fatalf("signup: %d", code)
	}
	a.idp.sub, a.idp.email = "a-1", "relay@privaterelay.appleid.com"
	if to := b.signIn(a.srv.URL, "apple", "?mode=link"); to != "/#/settings?connected=apple" {
		t.Fatalf("link went to %q", to)
	}
	if me := b.me(a.srv.URL); fmt.Sprint(me["linked"]) != "[apple]" || me["hasPassword"] != true {
		t.Fatalf("me after link: %v", me)
	}
	// Now "Continue with Apple" from a fresh browser signs in as ravi.
	b2 := a.browser()
	b2.signIn(a.srv.URL, "apple", "")
	if u := b2.me(a.srv.URL)["user"]; u != "ravi" {
		t.Fatalf("apple sign-in after link got %v", u)
	}
	// The same Apple account can't be linked to a second user.
	b3 := a.browser()
	b3.post(a.srv.URL, "/api/auth/signup", `{"username":"meera","password":"password123"}`)
	if to := b3.signIn(a.srv.URL, "apple", "?mode=link"); !strings.Contains(to, "already+connected") {
		t.Fatalf("double link went to %q", to)
	}
	// Linking requires being signed in.
	if to := a.browser().signIn(a.srv.URL, "apple", "?mode=link"); !strings.Contains(to, "Sign+in+first") {
		t.Fatalf("anonymous link went to %q", to)
	}
}

func TestSocialRejectsForgedCallbacks(t *testing.T) {
	a := newTestApp(t, "")
	a.idp.sub, a.idp.email = "g-5", "x@example.com"
	b := a.browser()
	// Start a real flow, then replay the callback with a different state.
	resp, _ := b.c.Get(a.srv.URL + "/api/auth/google/start")
	consent, _ := b.c.Get(resp.Header.Get("Location"))
	cb, _ := url.Parse(consent.Header.Get("Location"))
	q := cb.Query()
	q.Set("state", "attacker-state")
	cb.RawQuery = q.Encode()
	bad, _ := b.c.Get(cb.String())
	if loc := bad.Header.Get("Location"); !strings.Contains(loc, "auth_error=") {
		t.Fatalf("forged state went to %q", loc)
	}
	if u := b.me(a.srv.URL)["user"]; u != nil {
		t.Fatalf("forged callback signed in as %v", u)
	}
	// A callback with no state cookie at all (a different browser).
	other := a.browser()
	r2, _ := other.c.Get(consent.Header.Get("Location"))
	if loc := r2.Header.Get("Location"); !strings.Contains(loc, "expired") {
		t.Fatalf("callback without cookie went to %q", loc)
	}
	// The provider reporting "access_denied" (user pressed Cancel).
	r3, _ := b.c.Get(a.srv.URL + "/api/auth/google/callback?error=access_denied&state=x")
	if loc := r3.Header.Get("Location"); !strings.Contains(loc, "cancelled") {
		t.Fatalf("cancel went to %q", loc)
	}
	// Apple-style cross-site form POST to the callback is allowed through the
	// same-origin check (the state cookie protects it) and fails cleanly here.
	req, _ := http.NewRequest(http.MethodPost, a.srv.URL+"/api/auth/apple/callback", strings.NewReader("code=x&state=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://appleid.apple.com")
	r4, _ := b.c.Do(req)
	if r4.StatusCode != http.StatusFound || !strings.Contains(r4.Header.Get("Location"), "auth_error=") {
		t.Fatalf("cross-site POST callback: %s %q", r4.Status, r4.Header.Get("Location"))
	}
	// But other cross-site POSTs are still refused.
	req2, _ := http.NewRequest(http.MethodPost, a.srv.URL+"/api/auth/logout", nil)
	req2.Header.Set("Origin", "https://evil.test")
	if r5, _ := b.c.Do(req2); r5.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-site logout: %s", r5.Status)
	}
}

// TestSocialUIServer serves the whole app, wired to the fake provider, for a
// browser-driven test. Opt-in: LLDPREP_UI_TEST_ADDR=127.0.0.1:8095 (plus
// MONGODB_TEST_URI). It runs until LLDPREP_UI_TEST_SECONDS have passed.
func TestSocialUIServer(t *testing.T) {
	addr := os.Getenv("LLDPREP_UI_TEST_ADDR")
	if addr == "" {
		t.Skip("set LLDPREP_UI_TEST_ADDR to serve the app for a browser test")
	}
	if os.Getenv("MONGODB_TEST_URI") == "" {
		t.Skip("set MONGODB_TEST_URI too")
	}
	idp := newFakeProvider(t)
	idp.sub, idp.email = "g-ui", "priya@example.com"
	cs, err := content.Load(os.DirFS("content"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := accounts.Connect(context.Background(), os.Getenv("MONGODB_TEST_URI"), fmt.Sprintf("lldprep_ui_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.DropForTests(context.Background()); db.Close() })
	au := &auth{db: db, inviteCode: os.Getenv("LLDPREP_UI_TEST_INVITE"), limiter: newLoginLimiter(), social: &social{
		providers: map[string]*oauth.Provider{"google": idp.provider("google"), "apple": idp.provider("apple")},
		key:       []byte("test-key"),
	}}
	app := &app{content: cs, store: db, contentFS: os.DirFS("content")}
	mux := http.NewServeMux()
	app.routes(mux)
	au.routes(mux)
	// The fake provider lets the test pick who "consents" next.
	mux.HandleFunc("POST /test/next-identity", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		idp.mu.Lock()
		idp.sub, idp.email = r.Form.Get("sub"), r.Form.Get("email")
		idp.mu.Unlock()
	})
	mux.Handle("/", staticHandler(os.DirFS("web"), ""))
	srv := &http.Server{Addr: addr, Handler: au.requireUser(mux)}
	go srv.ListenAndServe()
	defer srv.Close()
	secs := 120
	fmt.Sscan(os.Getenv("LLDPREP_UI_TEST_SECONDS"), &secs)
	t.Logf("UI test server on http://%s for %ds", addr, secs)
	time.Sleep(time.Duration(secs) * time.Second)
}
