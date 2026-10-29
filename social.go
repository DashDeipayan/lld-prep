package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"lldprep/internal/accounts"
	"lldprep/internal/oauth"
)

// social is "Sign in with Google / Apple". It only exists in accounts mode.
type social struct {
	providers map[string]*oauth.Provider
	publicURL string // fixed origin for redirect URIs; empty = derive from the request
	key       []byte // signs the short-lived state cookie
}

const oauthCookie = "lldprep_oauth"

// oauthState travels in a signed cookie between /start and /callback. It binds
// the callback to the browser that started the flow (CSRF), and carries the
// nonce and PKCE verifier the provider's answer must match.
type oauthState struct {
	Provider string `json:"p"`
	State    string `json:"s"`
	Nonce    string `json:"n"`
	Verifier string `json:"v"`
	Mode     string `json:"m"`           // "login" or "link"
	Invited  bool   `json:"i,omitempty"` // a valid invite code was given (for sign-ups)
	LinkUser string `json:"u,omitempty"` // mode "link": the signed-in user to connect
	Expires  int64  `json:"e"`
}

// socialFromEnv configures whichever providers have credentials in the environment.
func socialFromEnv(publicURL string) (*social, error) {
	s := &social{providers: map[string]*oauth.Provider{}, publicURL: publicURL}
	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		secret := os.Getenv("GOOGLE_CLIENT_SECRET")
		if secret == "" {
			return nil, errors.New("GOOGLE_CLIENT_ID is set but GOOGLE_CLIENT_SECRET is not")
		}
		s.providers["google"] = oauth.NewGoogle(id, secret)
	}
	if id := os.Getenv("APPLE_CLIENT_ID"); id != "" {
		keyPEM := os.Getenv("APPLE_PRIVATE_KEY")
		if file := os.Getenv("APPLE_PRIVATE_KEY_FILE"); keyPEM == "" && file != "" {
			b, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}
			keyPEM = string(b)
		}
		p, err := oauth.NewApple(id, os.Getenv("APPLE_TEAM_ID"), os.Getenv("APPLE_KEY_ID"), keyPEM)
		if err != nil {
			return nil, err
		}
		s.providers["apple"] = p
	}
	if len(s.providers) == 0 {
		return nil, nil
	}
	if secret := os.Getenv("LLDPREP_SESSION_SECRET"); secret != "" {
		s.key = []byte(secret)
	} else {
		// A per-process key is enough: state cookies live for 10 minutes, so a
		// restart only interrupts sign-ins that are in progress at that moment.
		s.key = []byte(oauth.RandomString(32))
	}
	return s, nil
}

func (s *social) names() []string {
	out := []string{}
	if s == nil {
		return out
	}
	for name := range s.providers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (s *social) origin(r *http.Request) string {
	if s.publicURL != "" {
		return s.publicURL
	}
	scheme := "http"
	if isHTTPS(r) {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func (s *social) redirectURI(r *http.Request, provider string) string {
	return s.origin(r) + "/api/auth/" + provider + "/callback"
}

func (au *auth) socialRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/{provider}/start", au.handleOAuthStart)
	// Google redirects back with GET; Apple POSTs a form (response_mode=form_post).
	mux.HandleFunc("GET /api/auth/{provider}/callback", au.handleOAuthCallback)
	mux.HandleFunc("POST /api/auth/{provider}/callback", au.handleOAuthCallback)
}

// handleOAuthStart: /api/auth/google/start?mode=login|link&invite=CODE
func (au *auth) handleOAuthStart(w http.ResponseWriter, r *http.Request) {
	p := au.provider(r)
	if p == nil {
		http.NotFound(w, r)
		return
	}
	st := oauthState{
		Provider: p.Name, State: oauth.RandomString(24), Nonce: oauth.RandomString(24),
		Verifier: oauth.RandomString(48), Mode: "login", Expires: time.Now().Add(10 * time.Minute).Unix(),
	}
	if r.URL.Query().Get("mode") == "link" {
		user, ok := au.sessionUser(r)
		if !ok {
			authRedirect(w, r, "/", "Sign in first, then connect "+title(p.Name)+" from Settings.")
			return
		}
		st.Mode, st.LinkUser = "link", user
	}
	if invite := r.URL.Query().Get("invite"); invite != "" {
		if au.inviteCode != "" && !equalSecret(invite, au.inviteCode) {
			authRedirect(w, r, "/", "Wrong invite code.")
			return
		}
		st.Invited = true
	}
	au.setStateCookie(w, r, st)
	http.Redirect(w, r, p.AuthCodeURL(au.social.redirectURI(r, p.Name), st.State, st.Nonce, st.Verifier), http.StatusFound)
}

func (au *auth) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	p := au.provider(r)
	if p == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil { // covers both the query string (GET) and the form body (POST)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	st, ok := au.readStateCookie(r)
	clearStateCookie(w, r)
	back := "/"
	if ok && st.Mode == "link" {
		back = "/#/settings"
	}
	if e := r.Form.Get("error"); e != "" {
		msg := "Sign-in with " + title(p.Name) + " was cancelled."
		if e != "access_denied" && e != "user_cancelled_authorize" {
			msg = title(p.Name) + " sign-in failed: " + e
		}
		authRedirect(w, r, back, msg)
		return
	}
	if !ok || st.Provider != p.Name || !hmac.Equal([]byte(st.State), []byte(r.Form.Get("state"))) {
		authRedirect(w, r, back, "That sign-in link expired or didn't start here. Please try again.")
		return
	}
	id, err := p.Exchange(r.Context(), r.Form.Get("code"), au.social.redirectURI(r, p.Name), st.Verifier, st.Nonce)
	if err != nil {
		log.Printf("%s sign-in: %v", p.Name, err)
		authRedirect(w, r, back, title(p.Name)+" sign-in failed. Please try again.")
		return
	}
	ctx := r.Context()

	if st.Mode == "link" {
		user, signedIn := au.sessionUser(r)
		if !signedIn || user != st.LinkUser {
			authRedirect(w, r, "/", "Your session changed during sign-in. Please try again.")
			return
		}
		if err := au.db.LinkIdentity(ctx, id.Provider, id.Subject, user, id.Email); err != nil {
			msg := "Couldn't connect " + title(p.Name) + "."
			if errors.Is(err, accounts.ErrIdentityLinked) {
				msg = "That " + title(p.Name) + " account is already connected to a different user."
			}
			authRedirect(w, r, back, msg)
			return
		}
		http.Redirect(w, r, "/#/settings?connected="+p.Name, http.StatusFound)
		return
	}

	user, found, err := au.db.FindIdentity(ctx, id.Provider, id.Subject)
	if err != nil {
		httpError(w, err)
		return
	}
	if !found {
		// A new person. Sign-ups through Google/Apple obey the invite code too.
		if au.inviteCode != "" && !st.Invited {
			authRedirect(w, r, "/", "There's no account for that "+title(p.Name)+" login yet. Choose \"Create an account\", enter your invite code, then continue with "+title(p.Name)+".")
			return
		}
		hint := p.Name
		if at := strings.IndexByte(id.Email, '@'); at > 0 {
			hint = id.Email[:at]
		}
		if user, err = au.db.CreateExternalUser(ctx, hint); err == nil {
			err = au.db.LinkIdentity(ctx, id.Provider, id.Subject, user, id.Email)
		}
		if err != nil {
			httpError(w, err)
			return
		}
		log.Printf("new user %q signed up with %s", user, p.Name)
	}
	if err := au.issueSession(w, r, user); err != nil {
		httpError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (au *auth) provider(r *http.Request) *oauth.Provider {
	if !au.enabled() || au.social == nil {
		return nil
	}
	return au.social.providers[r.PathValue("provider")]
}

// authRedirect sends the browser back to the app with a message to show.
func authRedirect(w http.ResponseWriter, r *http.Request, to, msg string) {
	sep := "?"
	if !strings.Contains(to, "#") {
		to += "#/"
	}
	if strings.Contains(to, "?") {
		sep = "&"
	}
	http.Redirect(w, r, to+sep+"auth_error="+url.QueryEscape(msg), http.StatusFound)
}

// ---- the signed state cookie ----

func (au *auth) setStateCookie(w http.ResponseWriter, r *http.Request, st oauthState) {
	payload, _ := json.Marshal(st)
	body := base64.RawURLEncoding.EncodeToString(payload)
	http.SetCookie(w, &http.Cookie{
		Name: oauthCookie, Value: body + "." + au.sign(body), Path: "/api/auth/", MaxAge: 600,
		HttpOnly: true, Secure: isHTTPS(r), SameSite: stateSameSite(r),
	})
}

func (au *auth) readStateCookie(r *http.Request) (oauthState, bool) {
	var st oauthState
	c, err := r.Cookie(oauthCookie)
	if err != nil {
		return st, false
	}
	body, sig, found := strings.Cut(c.Value, ".")
	if !found || !hmac.Equal([]byte(sig), []byte(au.sign(body))) {
		return st, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil || json.Unmarshal(payload, &st) != nil || time.Now().Unix() > st.Expires {
		return st, false
	}
	return st, true
}

func clearStateCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: oauthCookie, Value: "", Path: "/api/auth/", MaxAge: -1,
		HttpOnly: true, Secure: isHTTPS(r), SameSite: stateSameSite(r)})
}

// Apple returns with a cross-site POST, which only carries cookies marked
// SameSite=None (and those must be Secure). Over plain http (local
// development), Lax is the only option that browsers accept.
func stateSameSite(r *http.Request) http.SameSite {
	if isHTTPS(r) {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (au *auth) sign(body string) string {
	m := hmac.New(sha256.New, au.social.key)
	m.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func title(provider string) string {
	switch provider {
	case "google":
		return "Google"
	case "apple":
		return "Apple"
	}
	return provider
}
