package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	slotuc "eventlineup/internal/usecase/slot"
	stageuc "eventlineup/internal/usecase/stage"
)

// stageRouterForOrg builds a stage router whose requests run as the given organizer.
func stageRouterForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(organizerID))
	httphandler.NewStageHandler(stageuc.New(database.NewStageRepository(pool))).Register(r.Group("/api"))
	return r
}

// slotRouterForOrg builds a slot router whose requests run as the given organizer.
func slotRouterForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(organizerID))
	httphandler.NewSlotHandler(slotuc.New(database.NewSlotRepository(pool))).Register(r.Group("/api"))
	return r
}

// EL-036: an organizer must not reach another organizer's stages, even though
// the event UUID is public (it is handed out via the public schedule link).
func TestStagesScopedToOrganizer(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	eventB := createTestEventForOrg(t, pool, orgB)
	stageB := createTestStage(t, pool, eventB)

	// Every request below is made by organizer A against organizer B's event.
	rA := stageRouterForOrg(t, pool, orgA)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list", http.MethodGet, "/api/events/" + eventB + "/stages", ""},
		{"get", http.MethodGet, "/api/events/" + eventB + "/stages/" + stageB, ""},
		{"create", http.MethodPost, "/api/events/" + eventB + "/stages", `{"name":"Intruder","color":"#FF0000"}`},
		{"patch", http.MethodPatch, "/api/events/" + eventB + "/stages/" + stageB, `{"name":"Hijacked","color":"#FF0000"}`},
		{"delete", http.MethodDelete, "/api/events/" + eventB + "/stages/" + stageB, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rA.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("organizer A on B's stage %s: expected 404, got %d: %s", tc.name, w.Code, w.Body.String())
			}
		})
	}

	// The list case deserves an explicit emptiness check: a 200 with B's stages
	// leaked would be just as bad as a write. Assert A sees no rows for B's event.
	t.Run("list leaks no rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventB+"/stages", nil)
		rA.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			var stages []map[string]any
			json.NewDecoder(w.Body).Decode(&stages)
			if len(stages) != 0 {
				t.Fatalf("organizer A must not see organizer B's stages, got %d", len(stages))
			}
		}
	})

	// Sanity: organizer B can still operate on their own stage.
	t.Run("owner still has access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventB+"/stages/"+stageB, nil)
		stageRouterForOrg(t, pool, orgB).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("owner should get 200 on their own stage, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// EL-036: an organizer must not reach another organizer's slots.
func TestSlotsScopedToOrganizer(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	eventB := createTestEventForOrg(t, pool, orgB)
	stageB := createTestStage(t, pool, eventB)
	slotB := createTestSlot(t, pool, eventB, stageB)

	rA := slotRouterForOrg(t, pool, orgA)
	writeBody := `{"stage_id":"` + stageB + `","slot_date":"2026-07-25","start_time":"18:00","end_time":"19:00"}`

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list", http.MethodGet, "/api/events/" + eventB + "/slots", ""},
		{"get", http.MethodGet, "/api/events/" + eventB + "/slots/" + slotB, ""},
		{"create", http.MethodPost, "/api/events/" + eventB + "/slots", writeBody},
		{"patch", http.MethodPatch, "/api/events/" + eventB + "/slots/" + slotB, writeBody},
		{"delete", http.MethodDelete, "/api/events/" + eventB + "/slots/" + slotB, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rA.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("organizer A on B's slot %s: expected 404, got %d: %s", tc.name, w.Code, w.Body.String())
			}
		})
	}

	t.Run("list leaks no rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventB+"/slots", nil)
		rA.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			var slots []map[string]any
			json.NewDecoder(w.Body).Decode(&slots)
			if len(slots) != 0 {
				t.Fatalf("organizer A must not see organizer B's slots, got %d", len(slots))
			}
		}
	})

	t.Run("owner still has access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventB+"/slots/"+slotB, nil)
		slotRouterForOrg(t, pool, orgB).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("owner should get 200 on their own slot, got %d: %s", w.Code, w.Body.String())
		}
	})
}
