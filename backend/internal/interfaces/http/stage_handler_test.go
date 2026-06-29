package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eventlineup/internal/domain/model"
)

func TestListStages(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)
	r := stageRouterForOrg(t, pool, org)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/stages", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stages []model.Stage
	if err := json.NewDecoder(w.Body).Decode(&stages); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestGetStage(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)
	r := stageRouterForOrg(t, pool, org)
	stageID := createTestStage(t, pool, eventID)

	t.Run("returns 200 with valid IDs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/stages/"+stageID, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var s model.Stage
		json.NewDecoder(w.Body).Decode(&s)
		if s.ID != stageID {
			t.Fatalf("expected id %s, got %s", stageID, s.ID)
		}
	})

	t.Run("returns 404 for unknown UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/stages/00000000-0000-0000-0000-000000000000", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestPatchStage(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)
	r := stageRouterForOrg(t, pool, org)
	stageID := createTestStage(t, pool, eventID)

	t.Run("updates name and color", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Main Stage", "color": "#FF0000"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/events/"+eventID+"/stages/"+stageID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var s model.Stage
		json.NewDecoder(w.Body).Decode(&s)
		if s.Name != "Main Stage" {
			t.Fatalf("expected name 'Main Stage', got %s", s.Name)
		}
		if s.Color != "#FF0000" {
			t.Fatalf("expected color '#FF0000', got %s", s.Color)
		}
	})

	t.Run("returns 404 for unknown stage UUID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "x", "color": "#000000"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/events/"+eventID+"/stages/00000000-0000-0000-0000-000000000000", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestCreateStage(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	eventID := createTestEventForOrg(t, pool, org)
	r := stageRouterForOrg(t, pool, org)
	body, _ := json.Marshal(map[string]string{"name": "A 主舞台", "color": "#6366F1"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/stages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var s model.Stage
	json.NewDecoder(w.Body).Decode(&s)
	if s.ID == "" {
		t.Fatal("expected ID in response")
	}
}
