package handler_test

import (
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

func authHandler(t *testing.T, secureCookies bool) *httphandler.AuthHandler {
	t.Helper()
	pool := setupTestDB(t)
	repo := database.NewOrganizerRepository(pool)
	g := googleauth.New("test-client-id", "test-secret", "http://localhost:8080/auth/google/callback")
	uc := authuc.New(g, repo, "auth-handler-test-secret-auth-handler-0123", time.Hour)
	return httphandler.NewAuthHandler(uc, "http://localhost:4200", secureCookies)
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
