package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GitHubAPI is the REST API base; tests point it at a fake server.
const GitHubAPI = "https://api.github.com"

// NewGitHub configures GitHub. Create an OAuth App under GitHub → Settings →
// Developer settings → OAuth Apps, with the callback URL
// <public url>/api/auth/github/callback.
func NewGitHub(clientID, clientSecret string) *Provider {
	return NewGitHubAt(clientID, clientSecret, "https://github.com", GitHubAPI)
}

// NewGitHubAt is NewGitHub for another GitHub host: GitHub Enterprise Server
// (web "https://ghe.example.com", api "https://ghe.example.com/api/v3") or a test server.
func NewGitHubAt(clientID, clientSecret, web, api string) *Provider {
	return &Provider{
		Name:         "github",
		ClientID:     clientID,
		AuthURL:      web + "/login/oauth/authorize",
		TokenURL:     web + "/login/oauth/access_token",
		Scopes:       []string{"read:user", "user:email"}, // read-only: profile and email addresses
		UsePKCE:      true,
		ClientSecret: func() (string, error) { return clientSecret, nil },
		Profile: func(ctx context.Context, c *http.Client, token string) (Identity, error) {
			return githubProfile(ctx, c, api, token)
		},
	}
}

// githubProfile asks GitHub who the token belongs to. The numeric user id is
// the stable identifier (logins can be renamed); the email is the primary,
// verified address, which may be private and so only visible via /user/emails.
func githubProfile(ctx context.Context, c *http.Client, api, token string) (Identity, error) {
	var user struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	if err := githubGet(ctx, c, api+"/user", token, &user); err != nil {
		return Identity{}, err
	}
	if user.ID == 0 {
		return Identity{}, fmt.Errorf("github: /user returned no id")
	}
	id := Identity{Subject: strconv.FormatInt(user.ID, 10), Username: user.Login}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	// The email is a nice-to-have (used for the username hint), so a failure here isn't fatal.
	if err := githubGet(ctx, c, api+"/user/emails", token, &emails); err == nil {
		for _, e := range emails {
			if e.Primary && e.Verified {
				id.Email, id.EmailVerified = strings.ToLower(e.Email), true
			}
		}
	}
	return id, nil
}

func githubGet(ctx context.Context, c *http.Client, url, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "lldprep") // GitHub rejects API requests without one
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, out)
}
