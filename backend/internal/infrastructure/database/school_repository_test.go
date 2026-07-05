package database_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/infrastructure/database"
)

// tenantTestPool gives the school/membership integration tests a live pool,
// skipping when DATABASE_URL is unset so unit runs stay green. Shared across the
// database_test package (used by membership_repository_test.go too).
func tenantTestPool(t *testing.T) *pgxpool.Pool {
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

// mkOrganizer creates a throwaway organizer (schools/memberships FK to it) and
// returns its id + email. The caller cleans up in FK order.
func mkOrganizer(t *testing.T, pool *pgxpool.Pool, key string) (id, email string) {
	t.Helper()
	email = key + "@tenant.test"
	o, err := database.NewOrganizerRepository(pool).Create(context.Background(), email, "Owner "+key, key)
	if err != nil {
		t.Fatalf("create organizer: %v", err)
	}
	return o.ID, email
}

func TestSchoolCreateAndGet(t *testing.T) {
	pool := tenantTestPool(t)
	ctx := context.Background()
	repo := database.NewSchoolRepository(pool)

	orgID, _ := mkOrganizer(t, pool, "school-create")
	created, err := repo.Create(ctx, "Test Academy", "Testville", orgID)
	if err != nil {
		t.Fatalf("create school: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM schools WHERE id = $1", created.ID)
		pool.Exec(ctx, "DELETE FROM organizers WHERE id = $1", orgID)
	})

	if created.ID == "" || created.Name != "Test Academy" || created.City != "Testville" || created.CreatedBy != orgID {
		t.Fatalf("unexpected school: %+v", created)
	}
	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get school: %v", err)
	}
	if got.ID != created.ID || got.Name != "Test Academy" {
		t.Fatalf("get mismatch: %+v", got)
	}
}

func TestSchoolGetNotFound(t *testing.T) {
	pool := tenantTestPool(t)
	_, err := database.NewSchoolRepository(pool).Get(context.Background(), "00000000-0000-0000-0000-0000000009ff")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
