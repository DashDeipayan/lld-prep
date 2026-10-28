package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"lldprep/internal/accounts"
)

// In local mode there are no accounts: everything belongs to this user.
const localUser = "local"

const sessionCookie = "lldprep_session"

type ctxKey struct{}

// userOf returns the signed-in user for a request that passed requireUser.
func userOf(r *http.Request) string {
	if u, ok := r.Context().Value(ctxKey{}).(string); ok {
		return u
	}
	return localUser
}

// auth holds the account-mode settings; a nil *accounts.Mongo means local mode.
type auth struct {
	db         *accounts.Mongo
	inviteCode string
	limiter    *loginLimiter
}

func (au *auth) enabled() bool { return au.db != nil }

func (au *auth) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/me", au.handleMe)
	mux.HandleFunc("POST /api/auth/signup", au.handleSignup)
	mux.HandleFunc("POST /api/auth/login", au.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", au.handleLogout)
}

// requireUser protects every /api/ route except /api/auth/*. In local mode it
// lets everything through as localUser.
func (au *auth) requireUser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !sameOrigin(r) {
			http.Error(w, "cross-origin request refused", http.StatusForbidden)
			return
		}
		if !au.enabled() || !strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/api/auth/") {
			h.ServeHTTP(w, r)
			return
		}
		user, ok := au.sessionUser(r)
		if !ok {
			http.Error(w, "sign in required", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, user)))
	})
}

func (au *auth) sessionUser(r *http.Request) (string, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return "", false
	}
	user, ok, err := au.db.SessionUser(r.Context(), c.Value)
	if err != nil {
		log.Printf("session lookup: %v", err)
		return "", false
	}
	return user, ok
}

func (au *auth) handleMe(w http.ResponseWriter, r *http.Request) {
	if !au.enabled() {
		writeJSON(w, map[string]any{"mode": "local", "user": localUser})
		return
	}
	out := map[string]any{"mode": "accounts", "inviteRequired": au.inviteCode != ""}
	if user, ok := au.sessionUser(r); ok {
		out["user"] = user
	}
	writeJSON(w, out)
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Invite   string `json:"invite"`
}

func readCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var c credentials
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<10)).Decode(&c); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return c, false
	}
	c.Username = accounts.NormalizeUsername(c.Username)
	if len(c.Password) > 72 { // bcrypt ignores anything longer
		http.Error(w, "password must be at most 72 characters", http.StatusBadRequest)
		return c, false
	}
	return c, true
}

func (au *auth) handleSignup(w http.ResponseWriter, r *http.Request) {
	if !au.enabled() {
		http.NotFound(w, r)
		return
	}
	c, ok := readCredentials(w, r)
	if !ok {
		return
	}
	if au.inviteCode != "" && !equalSecret(c.Invite, au.inviteCode) {
		http.Error(w, "wrong invite code", http.StatusForbidden)
		return
	}
	switch err := au.db.CreateUser(r.Context(), c.Username, c.Password); {
	case errors.Is(err, accounts.ErrUserExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case errors.Is(err, accounts.ErrInvalidUsername), errors.Is(err, accounts.ErrWeakPassword):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		httpError(w, err)
		return
	}
	au.startSession(w, r, c.Username)
}

func (au *auth) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !au.enabled() {
		http.NotFound(w, r)
		return
	}
	c, ok := readCredentials(w, r)
	if !ok {
		return
	}
	key := c.Username + "|" + clientIP(r)
	if !au.limiter.allow(key) {
		http.Error(w, "too many failed attempts; try again in a few minutes", http.StatusTooManyRequests)
		return
	}
	match, err := au.db.CheckPassword(r.Context(), c.Username, c.Password)
	if err != nil {
		httpError(w, err)
		return
	}
	if !match {
		au.limiter.fail(key)
		http.Error(w, "wrong username or password", http.StatusUnauthorized)
		return
	}
	au.limiter.reset(key)
	au.startSession(w, r, c.Username)
}

func (au *auth) startSession(w http.ResponseWriter, r *http.Request, user string) {
	token, exp, err := au.db.NewSession(r.Context(), user)
	if err != nil {
		httpError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", Expires: exp,
		HttpOnly: true, Secure: isHTTPS(r), SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, map[string]string{"user": user})
}

func (au *auth) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil && au.enabled() {
		if err := au.db.DeleteSession(r.Context(), c.Value); err != nil {
			log.Printf("logout: %v", err)
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: isHTTPS(r), SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

// isHTTPS is true for direct TLS and behind a TLS-terminating proxy (Render, Cloudflare).
func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// sameOrigin rejects state-changing requests sent by other sites. Browsers
// always send Origin on cross-site POST/PUT/DELETE.
func sameOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" {
		return true
	}
	u, err := url.Parse(o)
	return err == nil && u.Host == r.Host
}

func clientIP(r *http.Request) string {
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		return strings.TrimSpace(strings.Split(f, ",")[0])
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func equalSecret(got, want string) bool {
	g, wt := sha256.Sum256([]byte(got)), sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(g[:], wt[:]) == 1
}

// loginLimiter allows a few failed logins per username+IP per window.
type loginLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	fails  map[string][]time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{max: 10, window: 15 * time.Minute, fails: map[string][]time.Time{}}
}

func (l *loginLimiter) recent(key string, now time.Time) []time.Time {
	kept := l.fails[key][:0]
	for _, t := range l.fails[key] {
		if now.Sub(t) < l.window {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.fails, key)
		return nil
	}
	l.fails[key] = kept
	return kept
}

func (l *loginLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recent(key, time.Now())) < l.max
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.fails[key] = append(l.recent(key, now), now)
}

func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}
