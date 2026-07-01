package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	leaduc "eventlineup/internal/usecase/lead"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func leadRouter(t *testing.T, pool *pgxpool.Pool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	httphandler.NewLeadHandler(leaduc.New(database.NewLeadRepository(pool))).Register(r.Group("/api"))
	return r
}

func postLead(t *testing.T, r *gin.Engine, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/leads", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestCreateLead(t *testing.T) {
	pool := setupTestDB(t)
	r := leadRouter(t, pool)

	t.Run("valid lead → 201 and row persisted", func(t *testing.T) {
		w := postLead(t, r, map[string]any{
			"name": "Aria Chen", "organization": "Sub Bar Studio",
			"email": "aria@example.com", "message": "want a demo",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var got map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		id, _ := got["id"].(string)
		if id == "" {
			t.Fatalf("expected an id, got body %s", w.Body.String())
		}
		if got["source"] != "landing" {
			t.Fatalf("expected source=landing, got %v", got["source"])
		}
		t.Cleanup(func() { _, _ = pool.Exec(t.Context(), "DELETE FROM leads WHERE id=$1", id) })
	})

	t.Run("missing email → 400", func(t *testing.T) {
		w := postLead(t, r, map[string]any{"name": "No Email"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("malformed email → 400", func(t *testing.T) {
		w := postLead(t, r, map[string]any{"name": "Bad", "email": "not-an-email"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("name over 200 chars → 400", func(t *testing.T) {
		w := postLead(t, r, map[string]any{
			"name":  strings.Repeat("a", 201),
			"email": "toolong@example.com",
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("organization round-trips", func(t *testing.T) {
		w := postLead(t, r, map[string]any{
			"name":         "Orion Day",
			"email":        "orion@example.com",
			"organization": "Sub Bar Studio",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var got map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		id, _ := got["id"].(string)
		if id == "" {
			t.Fatalf("expected an id, got body %s", w.Body.String())
		}
		if got["organization"] != "Sub Bar Studio" {
			t.Fatalf("expected organization='Sub Bar Studio', got %v", got["organization"])
		}
		t.Cleanup(func() { _, _ = pool.Exec(t.Context(), "DELETE FROM leads WHERE id=$1", id) })
	})
}
