package oauth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strings"
	"sync"
	"time"
)

// appleSecret produces Apple's client secret: a JWT signed with ES256 using the
// developer's private key, valid for at most six months. We mint one that lasts
// a day and reuse it until it's close to expiring.
type appleSecret struct {
	clientID, teamID, keyID string
	key                     *ecdsa.PrivateKey

	mu      sync.Mutex
	token   string
	expires time.Time
}

func newAppleSecret(clientID, teamID, keyID, privateKeyPEM string) (*appleSecret, error) {
	if clientID == "" || teamID == "" || keyID == "" {
		return nil, errors.New("apple: client id, team id and key id are required")
	}
	// Env vars often arrive with literal "\n" instead of newlines.
	privateKeyPEM = strings.ReplaceAll(privateKeyPEM, `\n`, "\n")
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("apple: private key is not PEM (paste the whole .p8 file, including the BEGIN/END lines)")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("apple: private key is not a PKCS#8 key: " + err.Error())
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("apple: private key is not an EC key")
	}
	return &appleSecret{clientID: clientID, teamID: teamID, keyID: keyID, key: key}, nil
}

func (a *appleSecret) get() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Until(a.expires) > time.Hour {
		return a.token, nil
	}
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	header, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": a.keyID})
	claims, _ := json.Marshal(map[string]any{
		"iss": a.teamID,
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"aud": "https://appleid.apple.com",
		"sub": a.clientID,
	})
	signingInput := b64(header) + "." + b64(claims)
	digest := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, a.key, digest[:])
	if err != nil {
		return "", err
	}
	// JWS ES256 signatures are r || s, each left-padded to 32 bytes.
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	a.token, a.expires = signingInput+"."+b64(sig), exp
	return a.token, nil
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
