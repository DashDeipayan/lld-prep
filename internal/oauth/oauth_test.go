package oauth

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// fakeIdP is a minimal OpenID provider: a JWKS endpoint and a token endpoint
// that returns whatever ID token the test prepared.
type fakeIdP struct {
	key      *rsa.PrivateKey
	kid      string
	srv      *httptest.Server
	idToken  string
	lastForm url.Values
}

func newFakeIdP(t *testing.T) *fakeIdP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeIdP{key: key, kid: "k1"}
	mux := http.NewServeMux()
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "kid": f.kid, "alg": "RS256", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		f.lastForm = r.PostForm
		if r.PostForm.Get("code") != "good-code" {
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id_token": f.idToken})
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeIdP) sign(t *testing.T, kid string, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": kid, "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	input := b64(header) + "." + b64(payload)
	digest := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, f.key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return input + "." + b64(sig)
}

func (f *fakeIdP) provider() *Provider {
	return &Provider{
		Name: "google", ClientID: "client-123",
		AuthURL: f.srv.URL + "/auth", TokenURL: f.srv.URL + "/token", JWKSURL: f.srv.URL + "/keys",
		Issuers: []string{"https://issuer.test"}, Scopes: []string{"openid", "email"}, UsePKCE: true,
		ClientSecret: func() (string, error) { return "s3cret", nil },
	}
}

func goodClaims() map[string]any {
	return map[string]any{
		"iss": "https://issuer.test", "aud": "client-123", "sub": "user-42",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
		"nonce": "n-1", "email": "Asha@Example.com", "email_verified": true,
	}
}

func TestExchangeVerifiesAndReturnsIdentity(t *testing.T) {
	f := newFakeIdP(t)
	f.idToken = f.sign(t, f.kid, goodClaims())
	p := f.provider()
	id, err := p.Exchange(context.Background(), "good-code", "https://app.test/cb", "verifier-xyz", "n-1")
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if id.Subject != "user-42" || id.Email != "asha@example.com" || !id.EmailVerified || id.Provider != "google" {
		t.Fatalf("unexpected identity %+v", id)
	}
	if f.lastForm.Get("code_verifier") != "verifier-xyz" || f.lastForm.Get("client_secret") != "s3cret" ||
		f.lastForm.Get("redirect_uri") != "https://app.test/cb" {
		t.Fatalf("token request form: %v", f.lastForm)
	}
}

func TestVerifyRejectsBadTokens(t *testing.T) {
	f := newFakeIdP(t)
	p := f.provider()
	ctx := context.Background()
	mutate := func(k string, v any) string {
		c := goodClaims()
		if v == nil {
			delete(c, k)
		} else {
			c[k] = v
		}
		return f.sign(t, f.kid, c)
	}
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	forged := func() string {
		g := *f
		g.key = other
		return g.sign(t, f.kid, goodClaims())
	}()
	cases := map[string]string{
		"wrong issuer":   mutate("iss", "https://evil.test"),
		"wrong audience": mutate("aud", "someone-else"),
		"expired":        mutate("exp", time.Now().Add(-2*time.Hour).Unix()),
		"wrong nonce":    mutate("nonce", "replayed"),
		"no subject":     mutate("sub", ""),
		"forged":         forged,
		"unknown key":    f.sign(t, "k-unknown", goodClaims()),
		"garbage":        "not.a.jwt",
	}
	for name, tok := range cases {
		if _, err := p.VerifyIDToken(ctx, tok, "n-1"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	// audience as an array, email_verified as Apple's string form
	c := goodClaims()
	c["aud"] = []string{"x", "client-123"}
	c["email_verified"] = "true"
	id, err := p.VerifyIDToken(ctx, f.sign(t, f.kid, c), "n-1")
	if err != nil || !id.EmailVerified {
		t.Fatalf("array audience / string email_verified: %v %+v", err, id)
	}
}

func TestExchangeSurfacesTokenErrors(t *testing.T) {
	f := newFakeIdP(t)
	if _, err := f.provider().Exchange(context.Background(), "bad-code", "https://app.test/cb", "v", "n-1"); err == nil ||
		!strings.Contains(err.Error(), "invalid_grant") {
		t.Fatalf("want invalid_grant error, got %v", err)
	}
}

func TestAuthCodeURL(t *testing.T) {
	f := newFakeIdP(t)
	u, _ := url.Parse(f.provider().AuthCodeURL("https://app.test/cb", "st", "no", "verifier"))
	q := u.Query()
	if q.Get("state") != "st" || q.Get("nonce") != "no" || q.Get("code_challenge") != PKCEChallenge("verifier") ||
		q.Get("code_challenge_method") != "S256" || q.Get("scope") != "openid email" || q.Get("redirect_uri") != "https://app.test/cb" {
		t.Fatalf("auth url query: %v", q)
	}
}

func TestAppleClientSecretIsValidES256(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(key)
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	// Simulate an env var with literal \n sequences.
	p, err := NewApple("com.example.web", "TEAM123456", "KEY123", strings.ReplaceAll(pemKey, "\n", `\n`))
	if err != nil {
		t.Fatal(err)
	}
	tok, err := p.ClientSecret()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	var header map[string]string
	var claims map[string]any
	decodeSegment(parts[0], &header)
	decodeSegment(parts[1], &claims)
	if header["alg"] != "ES256" || header["kid"] != "KEY123" {
		t.Fatalf("header %v", header)
	}
	if claims["iss"] != "TEAM123456" || claims["sub"] != "com.example.web" || claims["aud"] != "https://appleid.apple.com" {
		t.Fatalf("claims %v", claims)
	}
	sig, _ := base64.RawURLEncoding.DecodeString(parts[2])
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r, s := new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])
	if len(sig) != 64 || !ecdsa.Verify(&key.PublicKey, digest[:], r, s) {
		t.Fatal("client secret signature does not verify")
	}
	again, _ := p.ClientSecret()
	if again != tok {
		t.Fatal("client secret should be cached")
	}
	if _, err := NewApple("a", "b", "c", "not a key"); err == nil {
		t.Fatal("expected an error for a bad key")
	}
}
