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
	"eventlineup/internal/interfaces/http/middleware"
	challengeuc "eventlineup/internal/usecase/challenge"
)

// --- shared student-feature test helpers (used by challenge + submission tests) ---

// studentContext simulates the StudentAuth middleware: it stashes the student's
// DJ id and school id, the keys the real middleware sets from a student JWT.
func studentContext(schoolID, djID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.StudentIDKey, djID)
		c.Set(middleware.SchoolIDKey, schoolID)
		c.Next()
	}
}

func schoolIDForOrg(t *testing.T, pool *pgxpool.Pool, org string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`SELECT school_id FROM memberships WHERE organizer_id = $1 AND status = 'active' LIMIT 1`,
		org).Scan(&id); err != nil {
		t.Fatalf("schoolIDForOrg: %v", err)
	}
	return id
}

func jsonReq(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func challengeTeacherRouter(pool *pgxpool.Pool, org string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(pool, org))
	httphandler.NewChallengeHandler(challengeuc.New(database.NewChallengeRepository(pool))).
		RegisterProtected(r.Group("/api"))
	return r
}

func challengeStudentRouter(pool *pgxpool.Pool, schoolID, djID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(studentContext(schoolID, djID))
	httphandler.NewChallengeHandler(challengeuc.New(database.NewChallengeRepository(pool))).
		RegisterStudent(r.Group("/api"))
	return r
}

func TestCreateAndListCoreChallenge(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := challengeTeacherRouter(pool, org)

	w := jsonReq(t, r, http.MethodPost, "/api/challenges", map[string]any{
		"type": "core", "title": "First Blend", "description": "blend two tracks",
		"badge_label": "First Blend", "badge_icon": "🎧", "sort_order": 10,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create core: want 201, got %d (%s)", w.Code, w.Body.String())
	}
	var created model.Challenge
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.ID == "" || created.Type != "core" || created.BadgeLabel != "First Blend" {
		t.Fatalf("unexpected created challenge: %+v", created)
	}

	// Teacher list includes it.
	w = jsonReq(t, r, http.MethodGet, "/api/challenges", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: got %d", w.Code)
	}
	var list []model.Challenge
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("teacher list should contain the challenge, got %+v", list)
	}

	// Student "browse core" includes it.
	sr := challengeStudentRouter(pool, schoolIDForOrg(t, pool, org), seedDJ(t, pool, org, "student-a"))
	w = jsonReq(t, sr, http.MethodGet, "/api/student/challenges/core", nil)
	var core []model.Challenge
	json.Unmarshal(w.Body.Bytes(), &core)
	if len(core) != 1 || core[0].Type != "core" {
		t.Fatalf("browse-core should list the core challenge, got %+v", core)
	}
}

func TestCreateCustomChallengeShowsForAssignedStudent(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	dj := seedDJ(t, pool, org, "assigned-student")
	r := challengeTeacherRouter(pool, org)

	w := jsonReq(t, r, http.MethodPost, "/api/challenges", map[string]any{
		"type": "custom", "title": "EQ transitions", "description": "work on EQ",
		"assigned_dj_id": dj, "sort_order": 5,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create custom: want 201, got %d (%s)", w.Code, w.Body.String())
	}

	sr := challengeStudentRouter(pool, schoolIDForOrg(t, pool, org), dj)
	w = jsonReq(t, sr, http.MethodGet, "/api/student/challenges", nil)
	var mine []model.Challenge
	json.Unmarshal(w.Body.Bytes(), &mine)
	found := false
	for _, c := range mine {
		if c.Title == "EQ transitions" {
			found = true
		}
	}
	if !found {
		t.Fatalf("assigned custom challenge should appear in student's list, got %+v", mine)
	}
}

func TestCreateChallengeRejectsInvalid(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := challengeTeacherRouter(pool, org)

	// core without badge → validation 400
	w := jsonReq(t, r, http.MethodPost, "/api/challenges", map[string]any{
		"type": "core", "title": "No Badge",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("core w/o badge: want 400, got %d", w.Code)
	}
}

func TestUpdateAndArchiveChallenge(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	r := challengeTeacherRouter(pool, org)

	w := jsonReq(t, r, http.MethodPost, "/api/challenges", map[string]any{
		"type": "core", "title": "Clean Cut", "badge_label": "Clean Cut", "badge_icon": "✂️",
	})
	var c model.Challenge
	json.Unmarshal(w.Body.Bytes(), &c)

	w = jsonReq(t, r, http.MethodPut, "/api/challenges/"+c.ID, map[string]any{
		"type": "core", "title": "Cleaner Cut", "badge_label": "Clean Cut", "badge_icon": "✂️",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update: want 200, got %d (%s)", w.Code, w.Body.String())
	}

	w = jsonReq(t, r, http.MethodDelete, "/api/challenges/"+c.ID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("archive: want 204, got %d", w.Code)
	}

	w = jsonReq(t, r, http.MethodGet, "/api/challenges", nil)
	var list []model.Challenge
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 0 {
		t.Fatalf("archived challenge should not list, got %+v", list)
	}
}
