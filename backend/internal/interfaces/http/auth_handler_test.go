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

func authHandler(t *testing.T) *httphandler.AuthHandler {
	t.Helper()
	pool := setupTestDB(t)
	repo := database.NewOrganizerRepository(pool)
	g := googleauth.New("test-client-id", "test-secret", "http://localhost:8080/auth/google/callback")
	uc := authuc.New(g, repo, "auth-handler-test-secret-auth-handler-0123", time.Hour)
	return httphandler.NewAuthHandler(uc, "http://localhost:4200")
}

func TestAuthGoogleRedirectsToGoogle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler(t).Register(r)

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
	if !strings.Contains(strings.Join(w.Header().Values("Set-Cookie"), ";"), "oauth_state=") {
		t.Fatal("expected oauth_state CSRF cookie to be set")
	}
}

func TestAuthCallbackMissingCodeReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler(t).Register(r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
