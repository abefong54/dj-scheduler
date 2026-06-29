package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/interfaces/http/middleware"
)

// captureLog redirects the standard logger to a buffer for the duration of the
// test and restores it afterward.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	flags, prefix, out := log.Flags(), log.Prefix(), log.Writer()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(out)
		log.SetFlags(flags)
		log.SetPrefix(prefix)
	})
	return &buf
}

func loggedEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.GET("/api/dj/portal", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// The whole point of EL-037: a secret carried in the query string must never
// reach the access log. We log the route template instead.
func TestRequestLoggerOmitsQueryString(t *testing.T) {
	buf := captureLog(t)
	r := loggedEngine()

	const secret = "supersecrettoken1234567890"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/dj/portal?token="+secret, nil))

	logged := buf.String()
	if strings.Contains(logged, secret) {
		t.Errorf("secret leaked into log line: %q", logged)
	}
	if strings.Contains(logged, "token=") {
		t.Errorf("query string leaked into log line: %q", logged)
	}
	if !strings.Contains(logged, "/api/dj/portal") {
		t.Errorf("expected route template in log, got %q", logged)
	}
}

// An unmatched path (404) must not leak its raw URL either — c.FullPath() is
// empty there, so we log a placeholder.
func TestRequestLoggerUnmatchedRouteHidesRawURL(t *testing.T) {
	buf := captureLog(t)
	r := loggedEngine()

	const secret = "leakytoken9876543210"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/google/callback?code="+secret, nil))

	logged := buf.String()
	if strings.Contains(logged, secret) {
		t.Errorf("secret leaked into log line for unmatched route: %q", logged)
	}
	if !strings.Contains(logged, "unmatched") {
		t.Errorf("expected 'unmatched' placeholder, got %q", logged)
	}
}
