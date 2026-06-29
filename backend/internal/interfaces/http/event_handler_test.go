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

	"eventlineup/internal/domain/model"
	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	eventuc "eventlineup/internal/usecase/event"
)

// eventRouterForOrg builds an event router whose requests run as the given organizer.
func eventRouterForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(organizerID))
	httphandler.NewEventHandler(eventuc.New(database.NewEventRepository(pool))).Register(r.Group("/api"))
	return r
}

func TestListEvents(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := eventRouterForOrg(t, pool, org)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []model.Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// US-002: a list returns only the caller's events, never another organizer's.
func TestListEventsScopedToOrganizer(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	eventA := createTestEventForOrg(t, pool, orgA)
	eventB := createTestEventForOrg(t, pool, orgB)

	r := eventRouterForOrg(t, pool, orgA)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []model.Event
	json.NewDecoder(w.Body).Decode(&events)

	ids := map[string]bool{}
	for _, e := range events {
		ids[e.ID] = true
	}
	if !ids[eventA] {
		t.Fatal("organizer A should see their own event")
	}
	if ids[eventB] {
		t.Fatal("organizer A must NOT see organizer B's event")
	}
}

func TestCreateEvent(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE name = 'test-create-event'")
	})
	r := eventRouterForOrg(t, pool, org)

	body, _ := json.Marshal(map[string]string{
		"name":       "test-create-event",
		"venue_name": "Test Club",
		"start_date": "2026-07-25",
		"end_date":   "2026-07-26",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var e model.Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID == "" {
		t.Fatal("expected ID in response")
	}

	// US-002: the new row is owned by the authenticated organizer.
	var ownerID string
	if err := pool.QueryRow(context.Background(),
		"SELECT organizer_id FROM events WHERE id = $1", e.ID).Scan(&ownerID); err != nil {
		t.Fatalf("read organizer_id: %v", err)
	}
	if ownerID != org {
		t.Fatalf("expected organizer_id %s, got %s", org, ownerID)
	}
}

func TestGetEvent(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	eventA := createTestEventForOrg(t, pool, orgA)

	t.Run("owner gets 200", func(t *testing.T) {
		r := eventRouterForOrg(t, pool, orgA)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events/"+eventA, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var e model.Event
		json.NewDecoder(w.Body).Decode(&e)
		if e.ID != eventA {
			t.Fatalf("expected id %s, got %s", eventA, e.ID)
		}
	})

	// US-002: cross-organizer access returns 404 (not 403 — don't leak existence).
	t.Run("other organizer gets 404", func(t *testing.T) {
		r := eventRouterForOrg(t, pool, orgB)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/events/"+eventA, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-organizer access, got %d", w.Code)
		}
	})
}

func TestDeleteEvent(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := eventRouterForOrg(t, pool, org)

	id := createTestEventForOrg(t, pool, org)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/events/"+id, nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil))
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestPatchEvent(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := eventRouterForOrg(t, pool, org)
	eventID := createTestEventForOrg(t, pool, org)

	t.Run("updates all mutable fields", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":       "updated-event",
			"venue_name": "Updated Venue",
			"start_date": "2026-08-01",
			"end_date":   "2026-08-02",
			"genres":     []string{"techno"},
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/events/"+eventID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var e model.Event
		json.NewDecoder(w.Body).Decode(&e)
		if e.Name != "updated-event" {
			t.Fatalf("expected name 'updated-event', got %s", e.Name)
		}
		if e.VenueName != "Updated Venue" {
			t.Fatalf("expected venue_name 'Updated Venue', got %s", e.VenueName)
		}
	})

	t.Run("returns 404 for unknown UUID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name": "x", "venue_name": "y", "start_date": "2026-08-01",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/events/00000000-0000-0000-0000-000000000000", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

// US-008: cloning copies the stage structure (not slots), prefixes the name with
// "Copy of ", and resets the dates to today so the organizer starts fresh.
func TestCloneEvent(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := eventRouterForOrg(t, pool, org)

	// Source event with one stage and one slot.
	id := createTestEventForOrg(t, pool, org)
	stageID := createTestStage(t, pool, id)
	createTestSlot(t, pool, id, stageID)

	clone := func(eventID string) (*httptest.ResponseRecorder, model.Event) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/clone", nil))
		var e model.Event
		json.NewDecoder(w.Body).Decode(&e)
		if e.ID != "" {
			t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", e.ID) })
		}
		return w, e
	}

	w, e := clone(id)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if e.ID == "" || e.ID == id {
		t.Fatalf("clone should have a new non-empty ID, got %q (source %q)", e.ID, id)
	}

	// Name is prefixed with "Copy of ".
	if e.Name != "Copy of test-event" {
		t.Fatalf(`expected name "Copy of test-event", got %q`, e.Name)
	}

	// Dates reset to today (CURRENT_DATE), not the source's past dates.
	var today string
	pool.QueryRow(context.Background(), "SELECT CURRENT_DATE::text").Scan(&today)
	if e.StartDate != today || e.EndDate != today {
		t.Fatalf("expected dates reset to today %s, got start=%s end=%s", today, e.StartDate, e.EndDate)
	}

	// Stages copied; slots NOT copied.
	var stageCount, slotCount int
	pool.QueryRow(context.Background(), "SELECT count(*) FROM stages WHERE event_id = $1", e.ID).Scan(&stageCount)
	pool.QueryRow(context.Background(), "SELECT count(*) FROM slots WHERE event_id = $1", e.ID).Scan(&slotCount)
	if stageCount != 1 {
		t.Fatalf("expected 1 stage copied, got %d", stageCount)
	}
	if slotCount != 0 {
		t.Fatalf("expected 0 slots on clone, got %d", slotCount)
	}

	// Cloning a clone works (no recursion / name issues): "Copy of Copy of test-event".
	w2, e2 := clone(e.ID)
	if w2.Code != http.StatusCreated {
		t.Fatalf("clone of clone: expected 201, got %d: %s", w2.Code, w2.Body.String())
	}
	if e2.Name != "Copy of Copy of test-event" {
		t.Fatalf(`expected "Copy of Copy of test-event", got %q`, e2.Name)
	}
}

// US-008 / US-002: cloning another organizer's event returns 404 (don't leak existence).
func TestCloneEventCrossOrganizer(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	eventA := createTestEventForOrg(t, pool, orgA)

	r := eventRouterForOrg(t, pool, orgB)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/events/"+eventA+"/clone", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-organizer clone, got %d: %s", w.Code, w.Body.String())
	}
}
