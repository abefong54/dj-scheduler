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

func orgTestPool(t *testing.T) *pgxpool.Pool {
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

func TestOrganizerCreateAndFindByGoogleID(t *testing.T) {
	pool := orgTestPool(t)
	repo := database.NewOrganizerRepository(pool)
	ctx := context.Background()

	const googleID = "test-google-id-create-find"
	pool.Exec(ctx, "DELETE FROM organizers WHERE google_id = $1", googleID)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM organizers WHERE google_id = $1", googleID) })

	created, err := repo.Create(ctx, "sso-test@example.com", "Test Organizer", googleID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated ID")
	}
	if created.Email != "sso-test@example.com" || created.Name != "Test Organizer" || created.GoogleID != googleID {
		t.Fatalf("unexpected created organizer: %+v", created)
	}

	found, err := repo.FindByGoogleID(ctx, googleID)
	if err != nil {
		t.Fatalf("FindByGoogleID: %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, found.ID)
	}
}

func TestOrganizerFindByGoogleIDNotFound(t *testing.T) {
	pool := orgTestPool(t)
	repo := database.NewOrganizerRepository(pool)

	_, err := repo.FindByGoogleID(context.Background(), "no-such-google-id")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
