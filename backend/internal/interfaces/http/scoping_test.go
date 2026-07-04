package handler_test

import (
	"bytes"
	"context"
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
	r.Use(orgContext(pool, organizerID))
	httphandler.NewStageHandler(stageuc.New(database.NewStageRepository(pool))).Register(r.Group("/api"))
	return r
}

// slotRouterForOrg builds a slot router whose requests run as the given organizer.
func slotRouterForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(pool, organizerID))
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

func createTestDJForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO djs (name, genre_tags, school_id)
		 VALUES ('Scoped DJ', '{}',
		         (SELECT school_id FROM memberships WHERE organizer_id = $1 AND status = 'active' LIMIT 1)) RETURNING id`,
		organizerID).Scan(&id); err != nil {
		t.Fatalf("createTestDJForOrg: %v", err)
	}
	return id
}

// EL-104: events are school-owned; a member of school A must not reach school B's
// event, even though event UUIDs are public.
func TestEventsScopedToSchool(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	eventB := createTestEventForOrg(t, pool, orgB)
	rA := eventRouterForOrg(t, pool, orgA)

	cases := []struct{ name, method, path, body string }{
		{"get", http.MethodGet, "/api/events/" + eventB, ""},
		{"patch", http.MethodPatch, "/api/events/" + eventB, `{"name":"Hijacked","venue_name":"X","start_date":"2026-07-25"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rA.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("school A on B's event %s: expected 404, got %d: %s", tc.name, w.Code, w.Body.String())
			}
		})
	}
	t.Run("list leaks no rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		rA.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events", nil))
		var events []map[string]any
		json.NewDecoder(w.Body).Decode(&events)
		for _, e := range events {
			if e["id"] == eventB {
				t.Fatalf("school A must not see school B's event")
			}
		}
	})
	t.Run("owner still has access", func(t *testing.T) {
		w := httptest.NewRecorder()
		eventRouterForOrg(t, pool, orgB).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events/"+eventB, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("owner should get 200 on their own event, got %d", w.Code)
		}
	})
}

// EL-104: DJs are school-owned; a member of school A must not reach school B's DJ.
func TestDJsScopedToSchool(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	djB := createTestDJForOrg(t, pool, orgB)
	rA := djRouterForOrg(t, pool, orgA)

	cases := []struct{ name, method, path, body string }{
		{"get", http.MethodGet, "/api/djs/" + djB, ""},
		{"patch", http.MethodPatch, "/api/djs/" + djB, `{"name":"Hijacked","genre_tags":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rA.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("school A on B's DJ %s: expected 404, got %d: %s", tc.name, w.Code, w.Body.String())
			}
		})
	}
	t.Run("list leaks no rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		rA.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs", nil))
		var djs []map[string]any
		json.NewDecoder(w.Body).Decode(&djs)
		for _, d := range djs {
			if d["id"] == djB {
				t.Fatalf("school A must not see school B's DJ")
			}
		}
	})
	t.Run("owner still has access", func(t *testing.T) {
		w := httptest.NewRecorder()
		djRouterForOrg(t, pool, orgB).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs/"+djB, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("owner should get 200 on their own DJ, got %d", w.Code)
		}
	})
}
