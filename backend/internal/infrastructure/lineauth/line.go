// Package lineauth implements the auth usecase's OAuthProvider against LINE
// Login (OAuth 2.1 / OIDC). Sign-in returns a LINE user id (sub) plus the
// profile carried in the verified id_token.
package lineauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	authuc "eventlineup/internal/usecase/auth"
)

const (
	authorizeURL = "https://access.line.me/oauth2/v2.1/authorize"
	tokenURL     = "https://api.line.me/oauth2/v2.1/token"
	// issuer is the `iss` LINE stamps into the id_token.
	issuer = "https://access.line.me"
)

type Authenticator struct {
	channelID     string
	channelSecret string
	redirectURL   string
	// Endpoints are fields (not consts) so tests can point them at a stub server.
	authorizeURL string
	tokenURL     string
	httpClient   *http.Client
}

func New(channelID, channelSecret, redirectURL string) *Authenticator {
	return &Authenticator{
		channelID:     channelID,
		channelSecret: channelSecret,
		redirectURL:   redirectURL,
		authorizeURL:  authorizeURL,
		tokenURL:      tokenURL,
		httpClient:    http.DefaultClient,
	}
}

// AuthCodeURL builds LINE's consent-screen URL. The email scope is required so
// the callback can link the account by email (see the auth usecase).
func (a *Authenticator) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", a.channelID)
	q.Set("redirect_uri", a.redirectURL)
	q.Set("state", state)
	q.Set("scope", "openid profile email")
	return a.authorizeURL + "?" + q.Encode()
}

// UserFromCode exchanges the auth code for a token set and returns the profile
// from the verified id_token. Secrets and tokens are never logged.
func (a *Authenticator) UserFromCode(ctx context.Context, code string) (authuc.ExternalUser, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", a.redirectURL)
	form.Set("client_id", a.channelID)
	form.Set("client_secret", a.channelSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return authuc.ExternalUser{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return authuc.ExternalUser{}, fmt.Errorf("line token exchange: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return authuc.ExternalUser{}, fmt.Errorf("line token status %d", resp.StatusCode)
	}

	var tok struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return authuc.ExternalUser{}, fmt.Errorf("decode line token: %w", err)
	}
	if tok.IDToken == "" {
		return authuc.ExternalUser{}, fmt.Errorf("line token response had no id_token")
	}
	return a.userFromIDToken(tok.IDToken)
}

// userFromIDToken verifies LINE's id_token — HMAC-signed with the channel secret,
// audience = channel id, issuer = access.line.me — and extracts the profile.
func (a *Authenticator) userFromIDToken(idToken string) (authuc.ExternalUser, error) {
	var claims struct {
		jwt.RegisteredClaims
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	_, err := jwt.ParseWithClaims(idToken, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected id_token signing method %v", t.Header["alg"])
		}
		return []byte(a.channelSecret), nil
	}, jwt.WithAudience(a.channelID), jwt.WithIssuer(issuer))
	if err != nil {
		return authuc.ExternalUser{}, fmt.Errorf("verify line id_token: %w", err)
	}
	return authuc.ExternalUser{
		Provider: authuc.ProviderLine,
		Subject:  claims.Subject,
		Email:    claims.Email,
		Name:     claims.Name,
	}, nil
}
