package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/infrastructure/database"
)

func dbPool(t *testing.T) *pgxpool.Pool {
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

// Second sign-in with the same Google ID must not create a duplicate organizer.
func TestUpsertOrganizerNoDuplicate(t *testing.T) {
	pool := dbPool(t)
	repo := database.NewOrganizerRepository(pool)
	uc := New(nil, repo, "usecase-test-secret-usecase-test-secret-01", time.Hour)
	ctx := context.Background()

	const gid = "test-google-id-upsert"
	pool.Exec(ctx, "DELETE FROM organizers WHERE google_id = $1", gid)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM organizers WHERE google_id = $1", gid) })

	gu := GoogleUser{GoogleID: gid, Email: "upsert@example.com", Name: "Upsert User"}

	first, err := uc.upsertOrganizer(ctx, gu)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	second, err := uc.upsertOrganizer(ctx, gu)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same organizer on re-signin, got %s vs %s", first.ID, second.ID)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM organizers WHERE google_id = $1", gid).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 organizer row, got %d", count)
	}
}
