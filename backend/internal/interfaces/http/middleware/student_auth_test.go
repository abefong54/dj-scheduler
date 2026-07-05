package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/token"
)

const mwTestSecret = "student-mw-secret-student-mw-secret-0123456789"

// runStudentAuth mounts StudentAuth on a one-route engine and returns the
// response plus whether the protected handler was reached.
func runStudentAuth(t *testing.T, authHeader string) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	reached := false
	r := gin.New()
	r.GET("/x", StudentAuth(mwTestSecret), func(c *gin.Context) {
		reached = true
		c.String(http.StatusOK, c.GetString(StudentIDKey)+"|"+c.GetString(SchoolIDKey))
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, reached
}

func TestStudentAuthAcceptsValidStudentToken(t *testing.T) {
	tok, _ := token.SignStudent(mwTestSecret, "dj-1", "school-1", time.Hour)
	w, reached := runStudentAuth(t, "Bearer "+tok)
	if w.Code != http.StatusOK || !reached {
		t.Fatalf("valid student token rejected: code=%d reached=%v", w.Code, reached)
	}
	if w.Body.String() != "dj-1|school-1" {
		t.Fatalf("context keys not set from token: %q", w.Body.String())
	}
}

func TestStudentAuthRejectsOrganizerToken(t *testing.T) {
	tok, _ := token.Sign(mwTestSecret, "org-1", "e@x.com", time.Hour) // role != student
	w, reached := runStudentAuth(t, "Bearer "+tok)
	if w.Code != http.StatusUnauthorized || reached {
		t.Fatalf("organizer token should be 401 and not reach handler: code=%d reached=%v", w.Code, reached)
	}
}

func TestStudentAuthRejectsMissingAndBadTokens(t *testing.T) {
	cases := map[string]string{
		"no header":     "",
		"not bearer":    "Token abc",
		"garbage token": "Bearer not-a-jwt",
		"wrong secret":  "Bearer " + mustSign(t, "other-secret-other-secret-other-secret-01", time.Hour),
		"expired token": "Bearer " + mustSign(t, mwTestSecret, -time.Hour),
	}
	for name, hdr := range cases {
		t.Run(name, func(t *testing.T) {
			w, reached := runStudentAuth(t, hdr)
			if w.Code != http.StatusUnauthorized || reached {
				t.Fatalf("%s: want 401 & not reached, got code=%d reached=%v", name, w.Code, reached)
			}
		})
	}
}

func mustSign(t *testing.T, secret string, ttl time.Duration) string {
	t.Helper()
	tok, err := token.SignStudent(secret, "dj-1", "school-1", ttl)
	if err != nil {
		t.Fatalf("SignStudent: %v", err)
	}
	return tok
}
