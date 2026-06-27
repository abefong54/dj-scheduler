// backend/handlers_djs_test.go
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

func djRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerDJRoutes(r.Group("/api"), pool)
	return r
}

func TestListDJs(t *testing.T) {
	pool := setupTestDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/djs", nil)
	djRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var djs []DJ
	if err := json.NewDecoder(w.Body).Decode(&djs); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateDJ(t *testing.T) {
	pool := setupTestDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM djs WHERE name = 'test-dj'")
	})

	body, _ := json.Marshal(map[string]interface{}{
		"name":       "test-dj",
		"genre_tags": []string{"hip hop", "r&b"},
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/djs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	djRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var d DJ
	json.NewDecoder(w.Body).Decode(&d)
	if d.ID == "" {
		t.Fatal("expected ID in response")
	}
	if len(d.GenreTags) != 2 {
		t.Fatalf("expected 2 genre tags, got %v", d.GenreTags)
	}
}
