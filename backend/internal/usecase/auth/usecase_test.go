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
	uc := New(nil, repo, database.NewMembershipRepository(pool), "usecase-test-secret-usecase-test-secret-01", time.Hour)
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

type stubGoogle struct{ user GoogleUser }

func (s stubGoogle) AuthCodeURL(string) string { return "" }
func (s stubGoogle) UserFromCode(context.Context, string) (GoogleUser, error) {
	return s.user, nil
}

// EL-104: signing in with an email that was invited to a school (a pending
// membership) links that membership to the new account and activates it.
func TestHandleCallbackLinksPendingInvite(t *testing.T) {
	pool := dbPool(t)
	ctx := context.Background()
	orgRepo := database.NewOrganizerRepository(pool)
	memRepo := database.NewMembershipRepository(pool)
	schoolRepo := database.NewSchoolRepository(pool)

	const ownerGID = "auth-link-owner-gid"
	const inviteeGID = "auth-link-invitee-gid"
	const inviteeEmail = "auth-link-invitee@example.com"
	pool.Exec(ctx, "DELETE FROM organizers WHERE google_id IN ($1,$2)", ownerGID, inviteeGID)

	owner, err := orgRepo.Create(ctx, "auth-link-owner@example.com", "Owner", ownerGID)
	if err != nil {
		t.Fatalf("owner: %v", err)
	}
	school, err := schoolRepo.Create(ctx, "Link School", "", owner.ID)
	if err != nil {
		t.Fatalf("school: %v", err)
	}
	if _, err := memRepo.Create(ctx, school.ID, nil, inviteeEmail, "member", "pending"); err != nil {
		t.Fatalf("pending membership: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM memberships WHERE school_id = $1", school.ID)
		pool.Exec(ctx, "DELETE FROM schools WHERE id = $1", school.ID)
		pool.Exec(ctx, "DELETE FROM organizers WHERE google_id IN ($1,$2)", ownerGID, inviteeGID)
	})

	uc := New(
		stubGoogle{user: GoogleUser{GoogleID: inviteeGID, Email: inviteeEmail, Name: "Invitee"}},
		orgRepo, memRepo, "auth-link-secret-auth-link-secret-000001", time.Hour,
	)
	if _, err := uc.HandleCallback(ctx, "code"); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}

	invitee, err := orgRepo.FindByGoogleID(ctx, inviteeGID)
	if err != nil {
		t.Fatalf("find invitee: %v", err)
	}
	m, err := memRepo.ActiveForOrganizer(ctx, invitee.ID)
	if err != nil {
		t.Fatalf("expected active membership after sign-in, got %v", err)
	}
	if m.SchoolID != school.ID || m.Status != "active" {
		t.Fatalf("invite not linked/activated: %+v", m)
	}
}
