package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/model"
)

// seedDJWithCerts inserts a DJ with the given certifications and student flag.
func seedDJWithCerts(t *testing.T, pool *pgxpool.Pool, organizerID, name string, certs []string, isStudent bool) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO djs (name, genre_tags, certifications, is_student, organizer_id)
		 VALUES ($1, '{}', $2, $3, $4) RETURNING id`,
		name, certs, isStudent, organizerID).Scan(&id); err != nil {
		t.Fatalf("seed dj with certs: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM djs WHERE id = $1", id) })
	return id
}

func listDJs(t *testing.T, r http.Handler, query string) []model.DJ {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs"+query, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list %q: expected 200, got %d: %s", query, w.Code, w.Body.String())
	}
	var djs []model.DJ
	if err := json.NewDecoder(w.Body).Decode(&djs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return djs
}

func ids(djs []model.DJ) map[string]bool {
	m := map[string]bool{}
	for _, d := range djs {
		m[d.ID] = true
	}
	return m
}

// EL-019: a brand-new DJ defaults to is_student=true with no certifications.
func TestNewDJDefaultsToStudentNoCerts(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, org, "fresh-dj")

	r := djRouterForOrg(t, pool, org)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/djs/"+djID, nil))
	var d model.DJ
	json.NewDecoder(w.Body).Decode(&d)

	if !d.IsStudent {
		t.Error("expected new DJ to default to is_student=true")
	}
	if len(d.Certifications) != 0 {
		t.Errorf("expected no certifications, got %v", d.Certifications)
	}
}

// EL-019: PATCH updates certifications + is_student and they persist.
func TestPatchDJCertifications(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, org, "cert-dj")
	r := djRouterForOrg(t, pool, org)

	body, _ := json.Marshal(map[string]any{
		"name":           "cert-dj",
		"genre_tags":     []string{},
		"certifications": []string{"House", "Hip Hop"},
		"is_student":     false,
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
	if len(d.Certifications) != 2 {
		t.Fatalf("expected 2 certifications, got %v", d.Certifications)
	}
	if d.IsStudent {
		t.Error("expected is_student=false after update")
	}

	// Persisted to the DB.
	var certs []string
	var isStudent bool
	pool.QueryRow(context.Background(),
		"SELECT certifications, is_student FROM djs WHERE id = $1", djID).Scan(&certs, &isStudent)
	if len(certs) != 2 || isStudent {
		t.Fatalf("DB not updated: certs=%v is_student=%v", certs, isStudent)
	}
}

// EL-019: GET /api/djs?certified_for=<genre> filters case-insensitively.
func TestListDJsCertifiedForFilter(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	houseDJ := seedDJWithCerts(t, pool, org, "house-dj", []string{"House"}, true)
	hipHopDJ := seedDJWithCerts(t, pool, org, "hiphop-dj", []string{"Hip Hop", "Pop"}, true)
	uncertDJ := seedDJWithCerts(t, pool, org, "uncert-dj", []string{}, true)

	r := djRouterForOrg(t, pool, org)
	// Case-insensitive: "hip hop" matches the stored "Hip Hop".
	got := ids(listDJs(t, r, "?certified_for=hip+hop"))
	if !got[hipHopDJ] {
		t.Error("expected hip-hop DJ in certified_for=hip+hop results")
	}
	if got[houseDJ] || got[uncertDJ] {
		t.Errorf("certified_for=hip+hop should exclude house/uncertified DJs, got %v", got)
	}
}

// EL-019: GET /api/djs?ready=true returns only DJs with at least one certification.
func TestListDJsReadyFilter(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	readyDJ := seedDJWithCerts(t, pool, org, "ready-dj", []string{"Techno"}, true)
	notReadyDJ := seedDJWithCerts(t, pool, org, "notready-dj", []string{}, true)

	r := djRouterForOrg(t, pool, org)
	got := ids(listDJs(t, r, "?ready=true"))
	if !got[readyDJ] {
		t.Error("expected certified DJ in ready=true results")
	}
	if got[notReadyDJ] {
		t.Error("ready=true must exclude DJs with no certifications")
	}
}
