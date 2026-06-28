package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	eventuc "eventlineup/internal/usecase/event"
	slotuc "eventlineup/internal/usecase/slot"
	stageuc "eventlineup/internal/usecase/stage"
)

func publicRouter(t *testing.T) *gin.Engine {
	t.Helper()
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httphandler.NewPublicHandler(
		eventuc.New(database.NewEventRepository(pool)),
		stageuc.New(database.NewStageRepository(pool)),
		slotuc.New(database.NewSlotRepository(pool)),
	)
	h.Register(r.Group("/api"))
	return r
}

type publicResponse struct {
	Event  model.Event   `json:"event"`
	Stages []model.Stage `json:"stages"`
	Slots  []model.Slot  `json:"slots"`
}

func TestPublicEventReturnsEventStagesSlots(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httphandler.NewPublicHandler(
		eventuc.New(database.NewEventRepository(pool)),
		stageuc.New(database.NewStageRepository(pool)),
		slotuc.New(database.NewSlotRepository(pool)),
	)
	h.Register(r.Group("/api"))

	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)
	createTestSlot(t, pool, eventID, stageID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/public", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body publicResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Event.ID != eventID {
		t.Fatalf("expected event id %s, got %s", eventID, body.Event.ID)
	}
	if len(body.Stages) < 1 {
		t.Fatalf("expected at least one stage, got %d", len(body.Stages))
	}
	if len(body.Slots) < 1 {
		t.Fatalf("expected at least one slot, got %d", len(body.Slots))
	}
}

func TestPublicEventUnknownIDReturns404(t *testing.T) {
	r := publicRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/00000000-0000-0000-0000-000000000000/public", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
