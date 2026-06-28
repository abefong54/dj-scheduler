package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/domain/model"
	httphandler "eventlineup/internal/interfaces/http"
	slotuc "eventlineup/internal/usecase/slot"
)

func TestListSlots(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/slots", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var slots []model.Slot
	if err := json.NewDecoder(w.Body).Decode(&slots); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("expected empty slot list, got %d", len(slots))
	}
}

func TestCreateSlot(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

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
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var s model.Slot
	json.NewDecoder(w.Body).Decode(&s)
	if s.ID == "" {
		t.Fatal("expected ID in response")
	}
}

func TestCreateSlot_StageOverlapReturns409(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)
	existingID := createTestSlot(t, pool, eventID, stageID) // 16:00–17:30 on stageID

	// New slot on the same stage and date overlapping 16:00–17:30.
	body, _ := json.Marshal(map[string]string{
		"stage_id":   stageID,
		"slot_date":  "2026-07-25",
		"start_time": "17:00",
		"end_time":   "18:00",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/slots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Error             string `json:"error"`
		Type              string `json:"type"`
		ConflictingSlotID string `json:"conflicting_slot_id"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "conflict" || resp.Type != "stage_overlap" {
		t.Fatalf("unexpected body: %+v", resp)
	}
	if resp.ConflictingSlotID != existingID {
		t.Fatalf("conflicting_slot_id = %s, want %s", resp.ConflictingSlotID, existingID)
	}
}

func TestCreateSlot_AdjacentReturns201(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)
	createTestSlot(t, pool, eventID, stageID) // 16:00–17:30 on stageID

	// Starts exactly when the existing slot ends — no overlap.
	body, _ := json.Marshal(map[string]string{
		"stage_id":   stageID,
		"slot_date":  "2026-07-25",
		"start_time": "17:30",
		"end_time":   "18:30",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/slots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetSlot(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)
	slotID := createTestSlot(t, pool, eventID, stageID)

	t.Run("returns 200 with valid IDs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/slots/"+slotID, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var s model.Slot
		json.NewDecoder(w.Body).Decode(&s)
		if s.ID != slotID {
			t.Fatalf("expected id %s, got %s", slotID, s.ID)
		}
	})

	t.Run("returns 404 for unknown UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/slots/00000000-0000-0000-0000-000000000000", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestUpdateSlot(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)
	slotID := createTestSlot(t, pool, eventID, stageID)

	t.Run("updates slot and returns updated fields", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"stage_id":   stageID,
			"slot_date":  "2026-07-25",
			"start_time": "16:00",
			"end_time":   "18:00",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/events/"+eventID+"/slots/"+slotID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var s model.Slot
		json.NewDecoder(w.Body).Decode(&s)
		if s.EndTime != "18:00" {
			t.Fatalf("expected end_time 18:00, got %s", s.EndTime)
		}
	})

	t.Run("returns 404 for unknown slot", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"stage_id":   stageID,
			"slot_date":  "2026-07-25",
			"start_time": "16:00",
			"end_time":   "17:30",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch,
			"/api/events/"+eventID+"/slots/00000000-0000-0000-0000-000000000000",
			bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestDeleteSlot(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewSlotRepository(pool)
	h := httphandler.NewSlotHandler(slotuc.New(repo))
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)

	t.Run("deletes slot and returns 204", func(t *testing.T) {
		slotID := createTestSlot(t, pool, eventID, stageID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/events/"+eventID+"/slots/"+slotID, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", w.Code)
		}
	})

	t.Run("returns 404 for unknown slot", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete,
			"/api/events/"+eventID+"/slots/00000000-0000-0000-0000-000000000000", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}
