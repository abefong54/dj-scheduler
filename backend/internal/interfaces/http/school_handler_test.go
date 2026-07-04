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

	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	schooluc "eventlineup/internal/usecase/school"
)

func schoolRouter(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(pool, organizerID))
	uc := schooluc.New(database.NewOrganizerRepository(pool), database.NewSchoolRepository(pool), database.NewMembershipRepository(pool))
	httphandler.NewSchoolHandler(uc).Register(r.Group("/api"))
	return r
}

// createRawOrganizer inserts an organizer with NO school/membership — the
// pre-onboarding state that createTestOrganizer (which auto-creates a school)
// can't represent.
func createRawOrganizer(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	var id string
	if err := pool.QueryRow(ctx,
		`INSERT INTO organizers (email, name, google_id)
		 VALUES (gen_random_uuid()::text||'@raw.local', 'Raw Org', gen_random_uuid()::text)
		 RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("createRawOrganizer: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM memberships WHERE organizer_id = $1", id)
		pool.Exec(ctx, "DELETE FROM schools WHERE created_by = $1", id)
		pool.Exec(ctx, "DELETE FROM organizers WHERE id = $1", id)
	})
	return id
}

func do(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestMeWithoutSchool(t *testing.T) {
	pool := setupTestDB(t)
	org := createRawOrganizer(t, pool)
	w := do(t, schoolRouter(t, pool, org), http.MethodGet, "/api/me", "")
	if w.Code != http.StatusOK {
		t.Fatalf("me: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["school"] != nil {
		t.Fatalf("expected school null pre-onboarding, got %v", resp["school"])
	}
	if resp["user"].(map[string]any)["id"] != org {
		t.Fatalf("me returned the wrong user")
	}
}

func TestRegisterSchoolThenMe(t *testing.T) {
	pool := setupTestDB(t)
	org := createRawOrganizer(t, pool)
	r := schoolRouter(t, pool, org)

	w := do(t, r, http.MethodPost, "/api/schools", `{"name":"iMDJ Academy","city":"Taipei"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var reg map[string]any
	json.NewDecoder(w.Body).Decode(&reg)
	if reg["name"] != "iMDJ Academy" || reg["role"] != "owner" {
		t.Fatalf("unexpected register response: %v", reg)
	}

	// /api/me re-resolves live, so the new school shows immediately.
	w2 := do(t, r, http.MethodGet, "/api/me", "")
	var me map[string]any
	json.NewDecoder(w2.Body).Decode(&me)
	school, _ := me["school"].(map[string]any)
	if school == nil || school["name"] != "iMDJ Academy" || school["role"] != "owner" {
		t.Fatalf("me should reflect the registered school, got %v", me["school"])
	}
}

func TestRegisterSchoolTwiceRejected(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool) // already owns a school
	w := do(t, schoolRouter(t, pool, org), http.MethodPost, "/api/schools", `{"name":"Second"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("registering a second school should be 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInviteMemberAsOwner(t *testing.T) {
	pool := setupTestDB(t)
	owner := createTestOrganizer(t, pool) // owner of a school
	r := schoolRouter(t, pool, owner)

	w := do(t, r, http.MethodPost, "/api/schools/current/members", `{"email":"teacher@imdj.test","role":"member"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("owner invite: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	// Duplicate email is rejected.
	if dup := do(t, r, http.MethodPost, "/api/schools/current/members", `{"email":"teacher@imdj.test","role":"member"}`); dup.Code != http.StatusBadRequest {
		t.Fatalf("duplicate invite should be 400, got %d", dup.Code)
	}
	// The invited member appears (pending) in the list.
	list := do(t, r, http.MethodGet, "/api/schools/current/members", "")
	var members []map[string]any
	json.NewDecoder(list.Body).Decode(&members)
	var found bool
	for _, m := range members {
		if m["email"] == "teacher@imdj.test" && m["status"] == "pending" {
			found = true
		}
	}
	if !found {
		t.Fatalf("invited member not found pending in list: %v", members)
	}
}

func TestInviteForbiddenWithoutOwnerRole(t *testing.T) {
	pool := setupTestDB(t)
	org := createRawOrganizer(t, pool) // no membership → no role
	w := do(t, schoolRouter(t, pool, org), http.MethodPost, "/api/schools/current/members", `{"email":"x@y.test","role":"member"}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("invite without owner/admin role should be 403, got %d: %s", w.Code, w.Body.String())
	}
}
