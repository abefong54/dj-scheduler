package lineauth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authuc "eventlineup/internal/usecase/auth"
)

func signIDToken(t *testing.T, secret, aud, iss, sub, email, name string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss":   iss,
		"aud":   aud,
		"sub":   sub,
		"email": email,
		"name":  name,
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign id_token: %v", err)
	}
	return s
}

func TestAuthCodeURL(t *testing.T) {
	a := New("chan-1", "sec", "https://app.example/auth/line/callback")
	u, err := url.Parse(a.AuthCodeURL("st8"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Host != "access.line.me" {
		t.Fatalf("host = %s, want access.line.me", u.Host)
	}
	q := u.Query()
	for k, want := range map[string]string{
		"response_type": "code",
		"client_id":     "chan-1",
		"redirect_uri":  "https://app.example/auth/line/callback",
		"state":         "st8",
		"scope":         "openid profile email",
	} {
		if q.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, q.Get(k), want)
		}
	}
}

func TestUserFromCode(t *testing.T) {
	const chID, chSecret = "channel-123", "channel-secret-xyz-channel-secret"
	idTok := signIDToken(t, chSecret, chID, issuer, "U-abc", "dj@line.me", "DJ Line")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if r.PostForm.Get("grant_type") != "authorization_code" || r.PostForm.Get("code") != "the-code" {
			t.Errorf("unexpected token request form: %v", r.PostForm)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"at","token_type":"Bearer","id_token":%q}`, idTok)
	}))
	defer srv.Close()

	a := New(chID, chSecret, "http://localhost:8080/auth/line/callback")
	a.tokenURL = srv.URL

	eu, err := a.UserFromCode(context.Background(), "the-code")
	if err != nil {
		t.Fatalf("UserFromCode: %v", err)
	}
	if eu.Provider != authuc.ProviderLine || eu.Subject != "U-abc" || eu.Email != "dj@line.me" || eu.Name != "DJ Line" {
		t.Fatalf("unexpected external user: %+v", eu)
	}
}

func TestUserFromCodeRejectsBadSignature(t *testing.T) {
	const chID, chSecret = "channel-123", "channel-secret-xyz-channel-secret"
	idTok := signIDToken(t, "WRONG-secret-WRONG-secret-WRONG!", chID, issuer, "U-abc", "dj@line.me", "DJ")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id_token":%q}`, idTok)
	}))
	defer srv.Close()

	a := New(chID, chSecret, "http://localhost:8080/auth/line/callback")
	a.tokenURL = srv.URL

	if _, err := a.UserFromCode(context.Background(), "code"); err == nil {
		t.Fatal("expected id_token signature verification to fail")
	}
}
