package database_test

import (
	"context"
	"errors"
	"testing"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/infrastructure/database"
)

// A LINE-first organizer is created with line_id and no google_id, and is found
// by line id + email but not by a google lookup.
func TestOrganizerCreateWithLineIDAndFind(t *testing.T) {
	pool := orgTestPool(t)
	repo := database.NewOrganizerRepository(pool)
	ctx := context.Background()

	const lineID = "test-line-id-create"
	const email = "line-create@example.com"
	pool.Exec(ctx, "DELETE FROM organizers WHERE line_id = $1", lineID)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM organizers WHERE line_id = $1", lineID) })

	created, err := repo.CreateWithLineID(ctx, email, "Line User", lineID)
	if err != nil {
		t.Fatalf("CreateWithLineID: %v", err)
	}
	if created.LineID != lineID || created.GoogleID != "" {
		t.Fatalf("expected line-only organizer (google_id empty), got %+v", created)
	}

	byLine, err := repo.FindByLineID(ctx, lineID)
	if err != nil || byLine.ID != created.ID {
		t.Fatalf("FindByLineID: err=%v org=%+v", err, byLine)
	}
	byEmail, err := repo.FindByEmail(ctx, email)
	if err != nil || byEmail.ID != created.ID {
		t.Fatalf("FindByEmail: err=%v org=%+v", err, byEmail)
	}
	if _, err := repo.FindByGoogleID(ctx, lineID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("a line-only user must not resolve by google id; got %v", err)
	}
}

// Linking attaches a line_id to an existing (Google) organizer — the same-email
// account-merge case.
func TestOrganizerLinkLineID(t *testing.T) {
	pool := orgTestPool(t)
	repo := database.NewOrganizerRepository(pool)
	ctx := context.Background()

	const googleID = "test-google-link-line"
	const lineID = "test-line-link"
	pool.Exec(ctx, "DELETE FROM organizers WHERE google_id = $1 OR line_id = $2", googleID, lineID)
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM organizers WHERE google_id = $1 OR line_id = $2", googleID, lineID)
	})

	org, err := repo.Create(ctx, "link@example.com", "Linker", googleID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.LinkLineID(ctx, org.ID, lineID); err != nil {
		t.Fatalf("LinkLineID: %v", err)
	}
	got, err := repo.FindByLineID(ctx, lineID)
	if err != nil {
		t.Fatalf("FindByLineID after link: %v", err)
	}
	if got.ID != org.ID || got.GoogleID != googleID || got.LineID != lineID {
		t.Fatalf("link mismatch: %+v", got)
	}
}

// The CHECK constraint forbids an organizer with neither provider identity.
func TestOrganizersRejectBothProvidersNull(t *testing.T) {
	pool := orgTestPool(t)
	ctx := context.Background()

	_, err := pool.Exec(ctx,
		"INSERT INTO organizers (email, name) VALUES ($1, $2)",
		"both-null@example.com", "No Provider")
	if err == nil {
		pool.Exec(ctx, "DELETE FROM organizers WHERE email = $1", "both-null@example.com")
		t.Fatal("expected CHECK to reject an organizer with neither google_id nor line_id")
	}
}
