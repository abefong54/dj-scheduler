package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/domain/model"
	httphandler "eventlineup/internal/interfaces/http"
	eventuc "eventlineup/internal/usecase/event"
)

func eventRouter(t *testing.T) *gin.Engine {
	t.Helper()
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewEventRepository(pool)
	h := httphandler.NewEventHandler(eventuc.New(repo))
	h.Register(r.Group("/api"))
	return r
}

func TestListEvents(t *testing.T) {
	r := eventRouter(t)
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

func TestCreateEvent(t *testing.T) {
	pool := setupTestDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE name = 'test-create-event'")
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewEventRepository(pool)
	h := httphandler.NewEventHandler(eventuc.New(repo))
	h.Register(r.Group("/api"))

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
}

func TestGetEvent(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewEventRepository(pool)
	h := httphandler.NewEventHandler(eventuc.New(repo))
	h.Register(r.Group("/api"))

	id := createTestEvent(t, pool)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var e model.Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID != id {
		t.Fatalf("expected id %s, got %s", id, e.ID)
	}
}

func TestDeleteEvent(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewEventRepository(pool)
	h := httphandler.NewEventHandler(eventuc.New(repo))
	h.Register(r.Group("/api"))

	id := createTestEvent(t, pool)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/events/"+id, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestPatchEvent(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewEventRepository(pool)
	h := httphandler.NewEventHandler(eventuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)

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

func TestDuplicateEvent(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewEventRepository(pool)
	h := httphandler.NewEventHandler(eventuc.New(repo))
	h.Register(r.Group("/api"))

	id := createTestEvent(t, pool)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+id+"/duplicate", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var e model.Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID == "" {
		t.Fatal("expected ID in duplicate response")
	}
	if e.ID == id {
		t.Fatal("duplicate should have a new ID")
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", e.ID)
	})
}
