package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/infrastructure/googleauth"
	httphandler "eventlineup/internal/interfaces/http"
	authuc "eventlineup/internal/usecase/auth"
)

// fakeLineProvider stands in for the real LINE provider so handler tests can
// drive the LINE routes without an external LINE round-trip. An empty email
// simulates a user who declined LINE's email permission.
type fakeLineProvider struct{ email string }

func (f fakeLineProvider) AuthCodeURL(state string) string {
	return "https://access.line.me/oauth2/v2.1/authorize?state=" + state
}
func (f fakeLineProvider) UserFromCode(context.Context, string) (authuc.ExternalUser, error) {
	return authuc.ExternalUser{Provider: authuc.ProviderLine, Subject: "U-x", Email: f.email, Name: "L"}, nil
}

func lineAuthHandler(t *testing.T, email string) *httphandler.AuthHandler {
	t.Helper()
	pool := setupTestDB(t)
	repo := database.NewOrganizerRepository(pool)
	providers := map[string]authuc.OAuthProvider{authuc.ProviderLine: fakeLineProvider{email: email}}
	uc := authuc.New(providers, repo, database.NewMembershipRepository(pool), "auth-handler-test-secret-auth-handler-0123", time.Hour)
	return httphandler.NewAuthHandler(uc, "http://localhost:4200", false, []string{authuc.ProviderLine})
}

func TestAuthLineRedirectsToLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	lineAuthHandler(t, "x@line.me").Register(r)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/line", nil))

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); !strings.HasPrefix(loc, "https://access.line.me/") {
		t.Fatalf("expected redirect to LINE, got %s", loc)
	}
}

// A LINE callback whose profile carries no email must not error out or create an
// account — it sends the user back to /login with a friendly reason.
func TestAuthLineCallbackNoEmailRedirectsToLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	lineAuthHandler(t, "").Register(r) // provider yields no email

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/line/callback?code=c&state=st", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "st"})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", w.Code, w.Body.String())
	}
	if loc := w.Header().Get("Location"); loc != "http://localhost:4200/login?error=no_email" {
		t.Fatalf("expected no_email login redirect, got %s", loc)
	}
}

func authHandler(t *testing.T, secureCookies bool) *httphandler.AuthHandler {
	t.Helper()
	pool := setupTestDB(t)
	repo := database.NewOrganizerRepository(pool)
	g := googleauth.New("test-client-id", "test-secret", "http://localhost:8080/auth/google/callback")
	providers := map[string]authuc.OAuthProvider{authuc.ProviderGoogle: g}
	uc := authuc.New(providers, repo, database.NewMembershipRepository(pool), "auth-handler-test-secret-auth-handler-0123", time.Hour)
	return httphandler.NewAuthHandler(uc, "http://localhost:4200", secureCookies, []string{authuc.ProviderGoogle})
}

func TestAuthGoogleRedirectsToGoogle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler(t, false).Register(r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://accounts.google.com/") {
		t.Fatalf("expected redirect to Google, got %s", loc)
	}
	cookie := strings.Join(w.Header().Values("Set-Cookie"), ";")
	if !strings.Contains(cookie, "oauth_state=") {
		t.Fatal("expected oauth_state CSRF cookie to be set")
	}
	// CSRF defenses: HttpOnly (no JS access) and SameSite=Lax (sent on the
	// top-level redirect back from Google, blocked on cross-site POSTs).
	if !strings.Contains(cookie, "HttpOnly") {
		t.Errorf("expected state cookie to be HttpOnly, got %q", cookie)
	}
	if !strings.Contains(cookie, "SameSite=Lax") {
		t.Errorf("expected state cookie SameSite=Lax, got %q", cookie)
	}
}

// With insecure cookies (local dev) the Secure attribute must be absent, or the
// browser drops the cookie over plain http and the OAuth flow breaks.
func TestAuthStateCookieNotSecureInDev(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler(t, false).Register(r)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/google", nil))

	if cookie := strings.Join(w.Header().Values("Set-Cookie"), ";"); strings.Contains(cookie, "Secure") {
		t.Errorf("expected no Secure attribute in dev, got %q", cookie)
	}
}

// In production (secureCookies=true) the state cookie must carry Secure so it is
// only ever sent over HTTPS.
func TestAuthStateCookieSecureInProd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler(t, true).Register(r)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/google", nil))

	if cookie := strings.Join(w.Header().Values("Set-Cookie"), ";"); !strings.Contains(cookie, "Secure") {
		t.Errorf("expected Secure attribute in prod, got %q", cookie)
	}
}

func TestAuthCallbackMissingCodeReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler(t, false).Register(r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
