// backend/handlers_stages_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func stageRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerStageRoutes(r.Group("/api"), pool)
	return r
}

func TestListStages(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/stages", nil)
	stageRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stages []Stage
	if err := json.NewDecoder(w.Body).Decode(&stages); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateStage(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)

	body, _ := json.Marshal(map[string]string{"name": "A 主舞台", "color": "#6366F1"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/stages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	stageRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var s Stage
	json.NewDecoder(w.Body).Decode(&s)
	if s.ID == "" {
		t.Fatal("expected ID in response")
	}
}
