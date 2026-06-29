package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	eventuc "eventlineup/internal/usecase/event"
	linenotifyuc "eventlineup/internal/usecase/linenotify"
)

// lineTestKey is a throwaway 32-byte AES key for the LINE token tests.
var lineTestKey = []byte("0123456789abcdef0123456789abcdef")

// lineRouterForOrg builds a router with both the event handler (for GET) and the
// LINE token handler, running as the given organizer.
func lineRouterForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(organizerID))
	api := r.Group("/api")
	repo := database.NewEventRepository(pool)
	httphandler.NewLineHandler(linenotifyuc.New(repo, lineTestKey)).Register(api)
	return r
}

func putLineToken(t *testing.T, r *gin.Engine, eventID, token string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"line_notify_token": token})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/events/"+eventID+"/line-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestSetLineTokenEnablesNotify(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)
	r := lineRouterForOrg(t, pool, org)

	w := putLineToken(t, r, eventID, "line-secret-token-abc123")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Enabled bool `json:"line_notify_enabled"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Enabled {
		t.Fatalf("expected line_notify_enabled true, got body %s", w.Body.String())
	}

	// Stored encrypted, never as plaintext.
	var enc *string
	pool.QueryRow(context.Background(),
		"SELECT line_notify_token_enc FROM events WHERE id=$1", eventID).Scan(&enc)
	if enc == nil {
		t.Fatal("expected line_notify_token_enc to be non-null")
	}
	if *enc == "line-secret-token-abc123" {
		t.Fatal("token stored in plaintext — must be encrypted at rest")
	}
}

func TestRemoveLineTokenDisablesNotify(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)
	r := lineRouterForOrg(t, pool, org)

	putLineToken(t, r, eventID, "some-token")
	w := putLineToken(t, r, eventID, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Enabled bool `json:"line_notify_enabled"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Enabled {
		t.Fatal("expected line_notify_enabled false after clearing the token")
	}

	var enc *string
	pool.QueryRow(context.Background(),
		"SELECT line_notify_token_enc FROM events WHERE id=$1", eventID).Scan(&enc)
	if enc != nil {
		t.Fatalf("expected line_notify_token_enc to be NULL after clearing, got %q", *enc)
	}
}

// AC: an event the organizer doesn't own → 403.
func TestSetLineTokenWrongOrganizerForbidden(t *testing.T) {
	pool := setupTestDB(t)
	owner := createTestOrganizer(t, pool)
	other := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, owner)

	r := lineRouterForOrg(t, pool, other) // acting as a different organizer
	w := putLineToken(t, r, eventID, "token")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-owner, got %d: %s", w.Code, w.Body.String())
	}
}

// AC: GET /api/events/:id exposes line_notify_enabled but never the raw token.
func TestGetEventExposesEnabledNotToken(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)

	// Build a combined router with both event GET and the line handler.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(org))
	api := r.Group("/api")
	repo := database.NewEventRepository(pool)
	httphandler.NewEventHandler(eventuc.New(repo)).Register(api)
	httphandler.NewLineHandler(linenotifyuc.New(repo, lineTestKey)).Register(api)

	const raw = "super-secret-line-token-zzz"
	putLineToken(t, r, eventID, raw)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events/"+eventID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"line_notify_enabled":true`) {
		t.Fatalf("expected line_notify_enabled true in GET body, got %s", body)
	}
	if strings.Contains(body, raw) {
		t.Fatalf("raw token leaked in GET response: %s", body)
	}
	if strings.Contains(body, "line_notify_token") && !strings.Contains(body, "line_notify_enabled") {
		t.Fatalf("response exposes a token field: %s", body)
	}
}
