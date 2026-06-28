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
	djuc "eventlineup/internal/usecase/dj"
)

func djRouterForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(organizerID))
	httphandler.NewDJHandler(djuc.New(database.NewDJRepository(pool))).Register(r.Group("/api"))
	return r
}

func seedDJ(t *testing.T, pool *pgxpool.Pool, organizerID, name string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO djs (name, genre_tags, organizer_id) VALUES ($1, '{"techno"}', $2) RETURNING id`,
		name, organizerID).Scan(&id); err != nil {
		t.Fatalf("seed dj: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM djs WHERE id = $1", id) })
	return id
}

func TestListDJs(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := djRouterForOrg(t, pool, org)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var djs []model.DJ
	if err := json.NewDecoder(w.Body).Decode(&djs); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// US-002: DJ roster is scoped to the organizer.
func TestListDJsScopedToOrganizer(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	djA := seedDJ(t, pool, orgA, "dj-a")
	djB := seedDJ(t, pool, orgB, "dj-b")

	r := djRouterForOrg(t, pool, orgA)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var djs []model.DJ
	json.NewDecoder(w.Body).Decode(&djs)

	ids := map[string]bool{}
	for _, d := range djs {
		ids[d.ID] = true
	}
	if !ids[djA] {
		t.Fatal("organizer A should see their own DJ")
	}
	if ids[djB] {
		t.Fatal("organizer A must NOT see organizer B's DJ")
	}
}

func TestCreateDJ(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM djs WHERE name = 'test-dj'")
	})
	r := djRouterForOrg(t, pool, org)

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

	var ownerID string
	if err := pool.QueryRow(context.Background(),
		"SELECT organizer_id FROM djs WHERE id = $1", d.ID).Scan(&ownerID); err != nil {
		t.Fatalf("read organizer_id: %v", err)
	}
	if ownerID != org {
		t.Fatalf("expected organizer_id %s, got %s", org, ownerID)
	}
}

func TestGetDJ(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, orgA, "get-test-dj")

	t.Run("owner gets 200", func(t *testing.T) {
		r := djRouterForOrg(t, pool, orgA)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs/"+djID, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var d model.DJ
		json.NewDecoder(w.Body).Decode(&d)
		if d.ID != djID {
			t.Fatalf("expected id %s, got %s", djID, d.ID)
		}
	})

	t.Run("other organizer gets 404", func(t *testing.T) {
		r := djRouterForOrg(t, pool, orgB)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs/"+djID, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-organizer access, got %d", w.Code)
		}
	})

	t.Run("returns 404 for unknown UUID", func(t *testing.T) {
		r := djRouterForOrg(t, pool, orgA)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs/00000000-0000-0000-0000-000000000000", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestPatchDJ(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := djRouterForOrg(t, pool, org)
	djID := seedDJ(t, pool, org, "patch-test-dj")

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
