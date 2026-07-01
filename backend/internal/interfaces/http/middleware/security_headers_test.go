package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/interfaces/http/middleware"
)

func securedEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	r.GET("/api/events", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/swagger/index.html", func(c *gin.Context) { c.String(http.StatusOK, "<html></html>") })
	return r
}

func TestSecurityHeadersOnAPIResponse(t *testing.T) {
	r := securedEngine()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events", nil))

	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "no-referrer",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'; base-uri 'none'",
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

// Swagger UI is an interactive HTML app; a default-src 'none' CSP would break it.
// The always-safe headers still apply, but the strict CSP must be omitted.
func TestSecurityHeadersExemptsSwaggerFromCSP(t *testing.T) {
	r := securedEngine()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))

	if got := w.Header().Get("Content-Security-Policy"); got != "" {
		t.Errorf("expected no CSP on swagger, got %q", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("expected nosniff on swagger, got %q", got)
	}
}
