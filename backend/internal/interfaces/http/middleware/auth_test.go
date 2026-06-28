package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-test-secret-test-secret-0123456789"

// mintToken signs a JWT with the given secret and expiry.
func mintToken(t *testing.T, secret, orgID string, exp time.Time) string {
	t.Helper()
	claims := Claims{
		OrganizerID: orgID,
		Email:       "organizer@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

// protectedRouter wires Auth in front of a handler that echoes the organizer ID
// and flips *called to true when reached.
func protectedRouter(secret string, called *bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Auth(secret))
	r.GET("/protected", func(c *gin.Context) {
		*called = true
		id, _ := c.Get(OrganizerIDKey)
		c.JSON(http.StatusOK, gin.H{"organizer_id": id})
	})
	return r
}

func TestAuthValidToken(t *testing.T) {
	called := false
	r := protectedRouter(testSecret, &called)

	token := mintToken(t, testSecret, "org-123", time.Now().Add(time.Hour))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["organizer_id"] != "org-123" {
		t.Fatalf("expected organizer_id 'org-123', got %q", body["organizer_id"])
	}
}

func TestAuthRejects(t *testing.T) {
	cases := []struct {
		name      string
		authHdr   func(t *testing.T) string // returns the Authorization header value, "" to omit
		wantError string
	}{
		{
			name:      "missing header",
			authHdr:   func(t *testing.T) string { return "" },
			wantError: "unauthorized",
		},
		{
			name:      "malformed header without Bearer scheme",
			authHdr:   func(t *testing.T) string { return mintToken(t, testSecret, "org-1", time.Now().Add(time.Hour)) },
			wantError: "unauthorized",
		},
		{
			name:      "expired token",
			authHdr:   func(t *testing.T) string { return "Bearer " + mintToken(t, testSecret, "org-1", time.Now().Add(-time.Hour)) },
			wantError: "token expired",
		},
		{
			name:      "token signed with wrong secret",
			authHdr:   func(t *testing.T) string { return "Bearer " + mintToken(t, "wrong-secret-wrong-secret-wrong-secret", "org-1", time.Now().Add(time.Hour)) },
			wantError: "unauthorized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			r := protectedRouter(testSecret, &called)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if hdr := tc.authHdr(t); hdr != "" {
				req.Header.Set("Authorization", hdr)
			}
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
			}
			if called {
				t.Fatal("handler must not be called on rejected request")
			}
			var body map[string]string
			json.NewDecoder(w.Body).Decode(&body)
			if body["error"] != tc.wantError {
				t.Fatalf("expected error %q, got %q", tc.wantError, body["error"])
			}
		})
	}
}
