// backend/handlers_slots_test.go
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

func slotRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerSlotRoutes(r.Group("/api"), pool)
	return r
}

func createTestStage(t *testing.T, pool *pgxpool.Pool, eventID string) string {
	t.Helper()
	var id string
	pool.QueryRow(context.Background(),
		`INSERT INTO stages (event_id, name, color) VALUES ($1, 'Test Stage', '#6366F1') RETURNING id`,
		eventID).Scan(&id)
	return id
}

func TestListSlots(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/slots", nil)
	slotRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var slots []Slot
	if err := json.NewDecoder(w.Body).Decode(&slots); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// new event has no slots
	if len(slots) != 0 {
		t.Fatalf("expected empty slot list, got %d", len(slots))
	}
}

func TestCreateSlot(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)

	body, _ := json.Marshal(map[string]string{
		"stage_id":   stageID,
		"slot_date":  "2026-07-25",
		"start_time": "16:00",
		"end_time":   "17:30",
		"notes":      "PA",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/slots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	slotRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var s Slot
	json.NewDecoder(w.Body).Decode(&s)
	if s.ID == "" {
		t.Fatal("expected ID in response")
	}
}
