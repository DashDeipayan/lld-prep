// Package oauth implements "Sign in with Google" and "Sign in with Apple" as
// OpenID Connect authorization-code flows, using only the standard library.
//
// The flow: AuthCodeURL sends the browser to the provider; the provider sends
// it back with a code; Exchange trades the code for an ID token at the
// provider's token endpoint and verifies that token's signature (against the
// provider's published keys), issuer, audience, expiry and nonce.
package oauth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

// Identity is who the provider says signed in.
type Identity struct {
	Provider      string
	Subject       string // stable, provider-scoped user id ("sub")
	Email         string
	EmailVerified bool
}

// Provider is one OpenID Connect identity provider.
type Provider struct {
	Name     string // "google" or "apple"
	ClientID string
	AuthURL  string
	TokenURL string
	JWKSURL  string
	Issuers  []string
	Scopes   []string
	// FormPost makes the provider POST the result back (Apple requires it when
	// asking for the email scope); otherwise it's a GET redirect.
	FormPost bool
	// UsePKCE adds a code_challenge (RFC 7636). Google supports it.
	UsePKCE bool
	// ClientSecret returns the client secret for the token request. Google's is
	// a fixed string; Apple's is a signed JWT generated on demand.
	ClientSecret func() (string, error)

	HTTP *http.Client
	keys *jwks
}

// NewGoogle configures Google. Create the OAuth client (type "Web application")
// in Google Cloud Console → APIs & Services → Credentials.
func NewGoogle(clientID, clientSecret string) *Provider {
	return &Provider{
		Name:         "google",
		ClientID:     clientID,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		JWKSURL:      "https://www.googleapis.com/oauth2/v3/certs",
		Issuers:      []string{"https://accounts.google.com", "accounts.google.com"},
		Scopes:       []string{"openid", "email"},
		UsePKCE:      true,
		ClientSecret: func() (string, error) { return clientSecret, nil },
	}
}

// NewApple configures Sign in with Apple. clientID is the Services ID; the
// private key is the .p8 key (PEM) created under Certificates, Identifiers &
// Profiles → Keys with "Sign in with Apple" enabled.
func NewApple(clientID, teamID, keyID, privateKeyPEM string) (*Provider, error) {
	secret, err := newAppleSecret(clientID, teamID, keyID, privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &Provider{
		Name:         "apple",
		ClientID:     clientID,
		AuthURL:      "https://appleid.apple.com/auth/authorize",
		TokenURL:     "https://appleid.apple.com/auth/token",
		JWKSURL:      "https://appleid.apple.com/auth/keys",
		Issuers:      []string{"https://appleid.apple.com"},
		Scopes:       []string{"email"},
		FormPost:     true,
		ClientSecret: secret.get,
	}, nil
}

func (p *Provider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// RandomString returns n random bytes, base64url-encoded: for state, nonce and PKCE verifiers.
func RandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // the OS random source failing is not recoverable
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// PKCEChallenge is the S256 code_challenge for a verifier.
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// AuthCodeURL is where to send the browser to start signing in.
func (p *Provider) AuthCodeURL(redirectURI, state, nonce, verifier string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {p.ClientID},
		"redirect_uri":  {redirectURI},
		"scope":         {strings.Join(p.Scopes, " ")},
		"state":         {state},
		"nonce":         {nonce},
	}
	if p.FormPost {
		q.Set("response_mode", "form_post")
	}
	if p.UsePKCE {
		q.Set("code_challenge", PKCEChallenge(verifier))
		q.Set("code_challenge_method", "S256")
	}
	if p.Name == "google" {
		q.Set("prompt", "select_account")
	}
	return p.AuthURL + "?" + q.Encode()
}

// Exchange trades an authorization code for a verified identity.
func (p *Provider) Exchange(ctx context.Context, code, redirectURI, verifier, nonce string) (Identity, error) {
	secret, err := p.ClientSecret()
	if err != nil {
		return Identity{}, err
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {p.ClientID},
		"client_secret": {secret},
	}
	if p.UsePKCE {
		form.Set("code_verifier", verifier)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return Identity{}, fmt.Errorf("%s token request: %w", p.Name, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return Identity{}, fmt.Errorf("%s token request: %s: %s", p.Name, resp.Status, strings.TrimSpace(string(body)))
	}
	var tok struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil || tok.IDToken == "" {
		return Identity{}, fmt.Errorf("%s token response has no id_token", p.Name)
	}
	return p.VerifyIDToken(ctx, tok.IDToken, nonce)
}

// idClaims are the ID token claims we use. Apple sends email_verified as a
// string ("true") and Google as a boolean, so it's decoded by hand.
type idClaims struct {
	Iss           string          `json:"iss"`
	Aud           json.RawMessage `json:"aud"` // a string, or an array of strings
	Sub           string          `json:"sub"`
	Exp           json.Number     `json:"exp"`
	Iat           json.Number     `json:"iat"`
	Nonce         string          `json:"nonce"`
	Email         string          `json:"email"`
	EmailVerified json.RawMessage `json:"email_verified"`
}

var errInvalidToken = errors.New("invalid ID token")

// VerifyIDToken checks the token's RS256 signature against the provider's
// published keys, then its issuer, audience, expiry and nonce.
func (p *Provider) VerifyIDToken(ctx context.Context, raw, nonce string) (Identity, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return Identity{}, errInvalidToken
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := decodeSegment(parts[0], &header); err != nil || header.Alg != "RS256" {
		return Identity{}, fmt.Errorf("%w: unsupported header", errInvalidToken)
	}
	if p.keys == nil {
		p.keys = &jwks{url: p.JWKSURL, client: p.client()}
	}
	key, err := p.keys.get(ctx, header.Kid)
	if err != nil {
		return Identity{}, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Identity{}, errInvalidToken
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		return Identity{}, fmt.Errorf("%w: bad signature", errInvalidToken)
	}

	var c idClaims
	dec := json.NewDecoder(strings.NewReader(mustSegment(parts[1])))
	dec.UseNumber()
	if err := dec.Decode(&c); err != nil {
		return Identity{}, fmt.Errorf("%w: claims", errInvalidToken)
	}
	if !slices.Contains(p.Issuers, c.Iss) {
		return Identity{}, fmt.Errorf("%w: issuer %q", errInvalidToken, c.Iss)
	}
	if !audienceContains(c.Aud, p.ClientID) {
		return Identity{}, fmt.Errorf("%w: audience", errInvalidToken)
	}
	exp, err := c.Exp.Int64()
	if err != nil || time.Now().After(time.Unix(exp, 0).Add(time.Minute)) { // 1 minute of clock skew
		return Identity{}, fmt.Errorf("%w: expired", errInvalidToken)
	}
	if nonce == "" || c.Nonce != nonce {
		return Identity{}, fmt.Errorf("%w: nonce mismatch", errInvalidToken)
	}
	if c.Sub == "" {
		return Identity{}, fmt.Errorf("%w: no subject", errInvalidToken)
	}
	verified := strings.Trim(string(c.EmailVerified), `"`) == "true"
	return Identity{Provider: p.Name, Subject: c.Sub, Email: strings.ToLower(c.Email), EmailVerified: verified}, nil
}

func decodeSegment(seg string, v any) error {
	b, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func mustSegment(seg string) string {
	b, _ := base64.RawURLEncoding.DecodeString(seg)
	return string(b)
}

func audienceContains(raw json.RawMessage, clientID string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == clientID
	}
	var many []string
	return json.Unmarshal(raw, &many) == nil && slices.Contains(many, clientID)
}

// ---- provider signing keys (JWKS), cached ----

type jwks struct {
	url     string
	client  *http.Client
	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
}

// get returns the key with this id, refetching the key set when it's unknown
// (providers rotate keys) or when the cache is older than an hour.
func (j *jwks) get(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if k, ok := j.keys[kid]; ok && time.Since(j.fetched) < time.Hour {
		return k, nil
	}
	if time.Since(j.fetched) > 10*time.Second { // don't hammer the endpoint with unknown kids
		if err := j.fetch(ctx); err != nil {
			return nil, err
		}
	}
	if k, ok := j.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("%w: unknown signing key %q", errInvalidToken, kid)
}

func (j *jwks) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, j.url, nil)
	if err != nil {
		return err
	}
	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching signing keys: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching signing keys: %s", resp.Status)
	}
	var set struct {
		Keys []struct {
			Kty, Kid, N, E string
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&set); err != nil {
		return fmt.Errorf("decoding signing keys: %w", err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		n, err1 := base64.RawURLEncoding.DecodeString(k.N)
		e, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	j.keys, j.fetched = keys, time.Now()
	return nil
}
