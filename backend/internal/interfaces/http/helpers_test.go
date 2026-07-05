package handler_test

import (
	"context"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/interfaces/http/middleware"
)

// orgContext is test middleware that simulates an authenticated caller: it sets
// the organizer ID and — resolving the organizer's active membership (EL-104) —
// their school ID, the same keys the real Auth + School middleware set. The
// school is resolved once at build time.
func orgContext(pool *pgxpool.Pool, organizerID string) gin.HandlerFunc {
	var schoolID, role string
	pool.QueryRow(context.Background(),
		`SELECT school_id, role FROM memberships WHERE organizer_id = $1 AND status = 'active' LIMIT 1`,
		organizerID).Scan(&schoolID, &role)
	return func(c *gin.Context) {
		c.Set(middleware.OrganizerIDKey, organizerID)
		if schoolID != "" {
			c.Set(middleware.SchoolIDKey, schoolID)
			c.Set(middleware.RoleKey, role)
		}
		c.Next()
	}
}

// createTestOrganizer inserts an organizer with unique credentials AND (EL-104) a
// school they own with an active owner membership, so scoped queries have a
// tenant to resolve. Returns the organizer ID. Distinct organizers get distinct
// schools, which is what makes the cross-tenant scoping tests exercise isolation.
func createTestOrganizer(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	var orgID, email string
	if err := pool.QueryRow(ctx,
		`INSERT INTO organizers (email, name, google_id)
		 VALUES (gen_random_uuid()::text||'@test.local', 'Test Organizer', gen_random_uuid()::text)
		 RETURNING id, email`).Scan(&orgID, &email); err != nil {
		t.Fatalf("createTestOrganizer: %v", err)
	}
	var schoolID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO schools (name, created_by) VALUES ('Test School', $1) RETURNING id`,
		orgID).Scan(&schoolID); err != nil {
		t.Fatalf("createTestOrganizer school: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO memberships (school_id, organizer_id, email, role, status)
		 VALUES ($1, $2, $3, 'owner', 'active')`, schoolID, orgID, email); err != nil {
		t.Fatalf("createTestOrganizer membership: %v", err)
	}
	t.Cleanup(func() {
		// FK-safe teardown of the whole tenant (events cascade to stages/slots).
		pool.Exec(ctx, "DELETE FROM djs WHERE school_id = $1", schoolID)
		pool.Exec(ctx, "DELETE FROM events WHERE school_id = $1", schoolID)
		pool.Exec(ctx, "DELETE FROM memberships WHERE school_id = $1", schoolID)
		pool.Exec(ctx, "DELETE FROM schools WHERE id = $1", schoolID)
		pool.Exec(ctx, "DELETE FROM organizers WHERE id = $1", orgID)
	})
	return orgID
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := database.InitDB(url)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// createTestEvent creates an event owned by a freshly created organizer's school.
func createTestEvent(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	return createTestEventForOrg(t, pool, createTestOrganizer(t, pool))
}

// createTestEventForOrg creates an event owned by the given organizer's school
// (EL-104: events carry both organizer_id (created_by) and school_id).
func createTestEventForOrg(t *testing.T, pool *pgxpool.Pool, organizerID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO events (name, venue_name, start_date, end_date, organizer_id, school_id)
		 VALUES ('test-event', 'Test Venue', '2026-07-25', '2026-07-26', $1,
		         (SELECT school_id FROM memberships WHERE organizer_id = $1 AND status = 'active' LIMIT 1))
		 RETURNING id`, organizerID).Scan(&id)
	if err != nil {
		t.Fatalf("createTestEvent: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", id)
	})
	return id
}

func createTestStage(t *testing.T, pool *pgxpool.Pool, eventID string) string {
	t.Helper()
	var id string
	pool.QueryRow(context.Background(),
		`INSERT INTO stages (event_id, name, color) VALUES ($1, 'Test Stage', '#6366F1') RETURNING id`,
		eventID).Scan(&id)
	return id
}

func createTestSlot(t *testing.T, pool *pgxpool.Pool, eventID, stageID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO slots (event_id, stage_id, slot_date, start_time, end_time)
		 VALUES ($1, $2, '2026-07-25', '16:00', '17:30') RETURNING id`,
		eventID, stageID).Scan(&id)
	if err != nil {
		t.Fatalf("createTestSlot: %v", err)
	}
	return id
}
