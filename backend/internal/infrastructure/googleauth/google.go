// Package googleauth implements the auth usecase's OAuthProvider against the
// real Google OAuth 2.0 endpoints.
package googleauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	authuc "eventlineup/internal/usecase/auth"
)

const userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

type Authenticator struct {
	cfg *oauth2.Config
}

func New(clientID, clientSecret, redirectURL string) *Authenticator {
	return &Authenticator{cfg: &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}}
}

func (a *Authenticator) AuthCodeURL(state string) string {
	return a.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// UserFromCode exchanges the OAuth code for a token and fetches the Google
// profile. Secrets and tokens are never logged.
func (a *Authenticator) UserFromCode(ctx context.Context, code string) (authuc.ExternalUser, error) {
	tok, err := a.cfg.Exchange(ctx, code)
	if err != nil {
		return authuc.ExternalUser{}, fmt.Errorf("oauth exchange failed: %w", err)
	}

	resp, err := a.cfg.Client(ctx, tok).Get(userInfoURL)
	if err != nil {
		return authuc.ExternalUser{}, fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return authuc.ExternalUser{}, fmt.Errorf("userinfo status %d", resp.StatusCode)
	}

	var info struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return authuc.ExternalUser{}, fmt.Errorf("decode userinfo: %w", err)
	}
	return authuc.ExternalUser{Provider: authuc.ProviderGoogle, Subject: info.ID, Email: info.Email, Name: info.Name}, nil
}
