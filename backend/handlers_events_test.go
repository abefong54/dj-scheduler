// backend/handlers_events_test.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func eventRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerEventRoutes(r.Group("/api"), pool)
	return r
}

// createTestEvent inserts a test event and registers cleanup.
func createTestEvent(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO events (name, venue_name, start_date, end_date)
		 VALUES ('test-event', 'Test Venue', '2026-07-25', '2026-07-26')
		 RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("createTestEvent: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", id)
	})
	return id
}

func TestListEvents(t *testing.T) {
	pool := setupTestDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateEvent(t *testing.T) {
	pool := setupTestDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE name = 'test-create-event'")
	})

	body, _ := json.Marshal(map[string]string{
		"name":       "test-create-event",
		"venue_name": "Test Club",
		"start_date": "2026-07-25",
		"end_date":   "2026-07-26",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var e Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID == "" {
		t.Fatal("expected ID in response")
	}
}

func TestGetEvent(t *testing.T) {
	pool := setupTestDB(t)
	id := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil)
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var e Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID != id {
		t.Fatalf("expected id %s, got %s", id, e.ID)
	}
}

func TestDeleteEvent(t *testing.T) {
	pool := setupTestDB(t)
	id := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/events/"+id, nil)
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
	// Verify it's gone
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil)
	eventRouter(pool).ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestDuplicateEvent(t *testing.T) {
	pool := setupTestDB(t)
	id := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+id+"/duplicate", nil)
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var e Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID == "" {
		t.Fatal("expected ID in duplicate response")
	}
	if e.ID == id {
		t.Fatal("duplicate should have a new ID")
	}
	// Cleanup the duplicate
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", e.ID)
	})
}
