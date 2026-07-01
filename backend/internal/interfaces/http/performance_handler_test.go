package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	perfuc "eventlineup/internal/usecase/performance"
)

// perfRouter builds a router exposing only the performance routes, authenticated
// as the given organizer (the same fake-middleware trick as the other handlers).
func perfRouter(t *testing.T, pool *pgxpool.Pool, organizerID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.Use(orgContext(organizerID))
	httphandler.NewPerformanceHandler(perfuc.New(database.NewPerformanceRepository(pool))).Register(api)
	return r
}

func getJSON(t *testing.T, r *gin.Engine, path string, out any) int {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	if w.Code == http.StatusOK && out != nil {
		if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatalf("decode %s: %v (body=%s)", path, err, w.Body.String())
		}
	}
	return w.Code
}

// seedDJ inserts a DJ for the organizer and cleans it up after the test.
func seedStudentDJ(t *testing.T, pool *pgxpool.Pool, organizerID, name string, isStudent bool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO djs (name, organizer_id, is_student) VALUES ($1, $2, $3) RETURNING id`,
		name, organizerID, isStudent).Scan(&id)
	if err != nil {
		t.Fatalf("seedDJ: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM djs WHERE id = $1", id) })
	return id
}

// seedSlot books a DJ into an event's stage. Slots are cleaned up via the event's
// ON DELETE CASCADE, so no explicit teardown is needed.
func seedSlot(t *testing.T, pool *pgxpool.Pool, eventID, stageID, djID, genre, date, start, end string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO slots (event_id, stage_id, dj_id, genre, slot_date, start_time, end_time)
		 VALUES ($1, $2, NULLIF($3,'')::uuid, $4, $5, $6, $7)`,
		eventID, stageID, djID, genre, date, start, end)
	if err != nil {
		t.Fatalf("seedSlot: %v", err)
	}
}

// seedReps books n one-hour House slots for a DJ on distinct days. Each DJ gets
// its own stage so concurrent reps across DJs don't trip the stage-overlap
// exclusion constraint.
func seedReps(t *testing.T, pool *pgxpool.Pool, eventID, djID string, n int) {
	t.Helper()
	stageID := createTestStage(t, pool, eventID)
	for i := 0; i < n; i++ {
		seedSlot(t, pool, eventID, stageID, djID, "House", fmt.Sprintf("2026-07-%02d", i+1), "20:00", "21:00")
	}
}

func TestDJPerformanceAggregatesAcrossEventsWithGenreBreakdown(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	ev := createTestEventForOrg(t, pool, org)
	stage := createTestStage(t, pool, ev)
	dj := seedStudentDJ(t, pool, org, "Cody", true)
	seedSlot(t, pool, ev, stage, dj, "House", "2026-07-01", "22:00", "23:30")  // 90 min
	seedSlot(t, pool, ev, stage, dj, "House", "2026-07-02", "20:00", "21:00")  // 60 min
	seedSlot(t, pool, ev, stage, dj, "Techno", "2026-07-03", "23:30", "00:30") // 60 min (past midnight)

	var p model.DJPerformance
	if code := getJSON(t, perfRouter(t, pool, org), "/api/djs/"+dj+"/performance", &p); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if p.Reps != 3 || p.TotalMinutes != 210 || p.LastPlayed != "2026-07-03" {
		t.Fatalf("aggregate: reps=%d minutes=%d last=%q, want 3/210/2026-07-03", p.Reps, p.TotalMinutes, p.LastPlayed)
	}
	byGenre := map[string]model.GenreStat{}
	for _, g := range p.ByGenre {
		byGenre[g.Genre] = g
	}
	if g := byGenre["House"]; g.Reps != 2 || g.TotalMinutes != 150 {
		t.Fatalf("House: %+v, want reps 2 / 150 min", g)
	}
	if g := byGenre["Techno"]; g.Reps != 1 || g.TotalMinutes != 60 {
		t.Fatalf("Techno: %+v, want reps 1 / 60 min", g)
	}
}

func TestDJPerformanceZeroRepOwnedDJ(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	dj := seedStudentDJ(t, pool, org, "NeverPlayed", true)

	var p model.DJPerformance
	if code := getJSON(t, perfRouter(t, pool, org), "/api/djs/"+dj+"/performance", &p); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if p.Reps != 0 || p.TotalMinutes != 0 || p.LastPlayed != "" {
		t.Fatalf("zero-rep: reps=%d minutes=%d last=%q, want 0/0/empty", p.Reps, p.TotalMinutes, p.LastPlayed)
	}
	if p.ByGenre == nil || len(p.ByGenre) != 0 {
		t.Fatalf("by_genre should be [] (got %#v)", p.ByGenre)
	}
}

func TestDJPerformanceOtherOrganizerIs404(t *testing.T) {
	pool := setupTestDB(t)
	org1 := createTestOrganizer(t, pool)
	org2 := createTestOrganizer(t, pool)
	djOfOrg2 := seedStudentDJ(t, pool, org2, "NotMine", true)

	if code := getJSON(t, perfRouter(t, pool, org1), "/api/djs/"+djOfOrg2+"/performance", nil); code != http.StatusNotFound {
		t.Fatalf("cross-organizer access: got %d, want 404", code)
	}
}

func TestRosterSummaryIncludesZeroRepStudentsExcludesGraduates(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	ev := createTestEventForOrg(t, pool, org)
	stage := createTestStage(t, pool, ev)
	played := seedStudentDJ(t, pool, org, "Played", true)
	zero := seedStudentDJ(t, pool, org, "ZeroRep", true)
	grad := seedStudentDJ(t, pool, org, "Graduate", false)
	seedSlot(t, pool, ev, stage, played, "House", "2026-07-01", "20:00", "21:00")
	seedSlot(t, pool, ev, stage, grad, "House", "2026-07-01", "21:00", "22:00")

	var roster []model.RosterPerformance
	if code := getJSON(t, perfRouter(t, pool, org), "/api/performance", &roster); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	byID := map[string]model.RosterPerformance{}
	for _, e := range roster {
		byID[e.DjID] = e
	}
	if e, ok := byID[played]; !ok || e.Reps != 1 {
		t.Fatalf("played student missing or wrong reps: %+v (ok=%v)", e, ok)
	}
	if e, ok := byID[zero]; !ok || e.Reps != 0 {
		t.Fatalf("zero-rep student must appear with reps 0: %+v (ok=%v)", e, ok)
	}
	if _, ok := byID[grad]; ok {
		t.Fatalf("graduate must not appear in the roster summary")
	}
}

func TestRosterSummaryWindowAndEventFilter(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	ev := createTestEventForOrg(t, pool, org)
	stage := createTestStage(t, pool, ev)
	dj := seedStudentDJ(t, pool, org, "Spanning", true)
	zero := seedStudentDJ(t, pool, org, "ZeroRep", true)
	seedSlot(t, pool, ev, stage, dj, "House", "2026-01-15", "20:00", "21:00") // out of window
	seedSlot(t, pool, ev, stage, dj, "House", "2026-07-15", "20:00", "21:00") // in window

	r := perfRouter(t, pool, org)

	// Window filter: only the July slot counts; the zero-rep student still appears.
	var windowed []model.RosterPerformance
	if code := getJSON(t, r, "/api/performance?from=2026-07-01&to=2026-07-31", &windowed); code != http.StatusOK {
		t.Fatalf("window status %d", code)
	}
	got := map[string]int{}
	for _, e := range windowed {
		got[e.DjID] = e.Reps
	}
	if got[dj] != 1 {
		t.Fatalf("windowed reps for spanning dj = %d, want 1", got[dj])
	}
	if r0, ok := got[zero]; !ok || r0 != 0 {
		t.Fatalf("zero-rep student must still appear in windowed roster with 0 reps (got %d, ok=%v)", r0, ok)
	}

	// event_id filter exercises the NULLIF($2,'')::uuid cast path.
	var byEvent []model.RosterPerformance
	if code := getJSON(t, r, "/api/performance?event_id="+ev, &byEvent); code != http.StatusOK {
		t.Fatalf("event filter status %d", code)
	}
	evReps := map[string]int{}
	for _, e := range byEvent {
		evReps[e.DjID] = e.Reps
	}
	if evReps[dj] != 2 {
		t.Fatalf("event-filtered reps = %d, want 2 (both slots in this event)", evReps[dj])
	}
}

func TestUnderservedThresholdAndGraduateExclusion(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	ev := createTestEventForOrg(t, pool, org)
	above := seedStudentDJ(t, pool, org, "Above", true) // 4 reps
	mid := seedStudentDJ(t, pool, org, "Mid", true)      // 2 reps
	zero := seedStudentDJ(t, pool, org, "Zero", true)    // 0 reps
	gradZero := seedStudentDJ(t, pool, org, "GradZero", false)
	seedReps(t, pool, ev, above, 4)
	seedReps(t, pool, ev, mid, 2)
	// gradZero has 0 reps but is a graduate → must be excluded even though < threshold.

	var rows []model.RosterPerformance
	if code := getJSON(t, perfRouter(t, pool, org), "/api/performance/underserved?threshold=3", &rows); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	in := map[string]bool{}
	for _, e := range rows {
		in[e.DjID] = true
	}
	if !in[mid] || !in[zero] {
		t.Fatalf("under-served (threshold 3) must include mid(2) and zero(0): %v", in)
	}
	if in[above] {
		t.Fatalf("above-threshold student (4 reps) must be excluded")
	}
	if in[gradZero] {
		t.Fatalf("graduate must be excluded from under-served even with 0 reps")
	}
}

func TestUnderservedDefaultThresholdWhenOmitted(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	ev := createTestEventForOrg(t, pool, org)
	below := seedStudentDJ(t, pool, org, "Below", true) // 2 reps < default(3)
	at := seedStudentDJ(t, pool, org, "At", true)        // 3 reps == default → excluded
	seedReps(t, pool, ev, below, 2)
	seedReps(t, pool, ev, at, 3)

	var rows []model.RosterPerformance
	if code := getJSON(t, perfRouter(t, pool, org), "/api/performance/underserved", &rows); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	in := map[string]bool{}
	for _, e := range rows {
		in[e.DjID] = true
	}
	if !in[below] {
		t.Fatalf("default threshold should flag the 2-rep student")
	}
	if in[at] {
		t.Fatalf("default threshold (%d) should exclude the 3-rep student", perfuc.DefaultUnderservedThreshold)
	}
}
