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
	httphandler "eventlineup/internal/interfaces/http"
	"eventlineup/internal/domain/model"
	djuc "eventlineup/internal/usecase/dj"
)

func djRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewDJRepository(pool)
	h := httphandler.NewDJHandler(djuc.New(repo))
	h.Register(r.Group("/api"))
	return r, func() { pool.Close() }
}

func TestListDJs(t *testing.T) {
	r, cleanup := djRouter(t)
	defer cleanup()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/djs", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var djs []model.DJ
	if err := json.NewDecoder(w.Body).Decode(&djs); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateDJ(t *testing.T) {
	pool := setupTestDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM djs WHERE name = 'test-dj'")
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewDJRepository(pool)
	h := httphandler.NewDJHandler(djuc.New(repo))
	h.Register(r.Group("/api"))

	body, _ := json.Marshal(map[string]interface{}{
		"name":       "test-dj",
		"genre_tags": []string{"hip hop", "r&b"},
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/djs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var d model.DJ
	json.NewDecoder(w.Body).Decode(&d)
	if d.ID == "" {
		t.Fatal("expected ID in response")
	}
	if len(d.GenreTags) != 2 {
		t.Fatalf("expected 2 genre tags, got %v", d.GenreTags)
	}
}

func TestGetDJ(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewDJRepository(pool)
	h := httphandler.NewDJHandler(djuc.New(repo))
	h.Register(r.Group("/api"))

	var djID string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO djs (name, genre_tags) VALUES ('get-test-dj', '{"techno"}') RETURNING id`,
	).Scan(&djID); err != nil {
		t.Fatalf("seed dj: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM djs WHERE id = $1", djID) })

	t.Run("returns 200 with valid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/djs/"+djID, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var d model.DJ
		json.NewDecoder(w.Body).Decode(&d)
		if d.ID != djID {
			t.Fatalf("expected id %s, got %s", djID, d.ID)
		}
	})

	t.Run("returns 404 for unknown UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/djs/00000000-0000-0000-0000-000000000000", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestPatchDJ(t *testing.T) {
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := database.NewDJRepository(pool)
	h := httphandler.NewDJHandler(djuc.New(repo))
	h.Register(r.Group("/api"))

	var djID string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO djs (name, genre_tags) VALUES ('patch-test-dj', '{}') RETURNING id`,
	).Scan(&djID); err != nil {
		t.Fatalf("seed dj: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM djs WHERE id = $1", djID) })

	t.Run("updates name and genre tags", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":       "updated-dj",
			"genre_tags": []string{"house", "techno"},
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/djs/"+djID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var d model.DJ
		json.NewDecoder(w.Body).Decode(&d)
		if d.Name != "updated-dj" {
			t.Fatalf("expected name 'updated-dj', got %s", d.Name)
		}
		if len(d.GenreTags) != 2 {
			t.Fatalf("expected 2 genre tags, got %v", d.GenreTags)
		}
	})

	t.Run("returns 404 for unknown UUID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"name": "x", "genre_tags": []string{}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/djs/00000000-0000-0000-0000-000000000000", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}
