package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/infrastructure/database"
)

// mkSchool creates an organizer + school and registers FK-ordered cleanup
// (memberships → schools → organizers). Returns the organizer id + email and the
// school id.
func mkSchool(t *testing.T, pool *pgxpool.Pool, key string) (orgID, orgEmail, schoolID string) {
	t.Helper()
	ctx := context.Background()
	orgID, orgEmail = mkOrganizer(t, pool, key)
	s, err := database.NewSchoolRepository(pool).Create(ctx, "School "+key, "", orgID)
	if err != nil {
		t.Fatalf("create school: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM memberships WHERE school_id = $1", s.ID)
		pool.Exec(ctx, "DELETE FROM schools WHERE id = $1", s.ID)
		pool.Exec(ctx, "DELETE FROM organizers WHERE id = $1", orgID)
	})
	return orgID, orgEmail, s.ID
}

func TestMembershipCreateAndActiveForOrganizer(t *testing.T) {
	pool := tenantTestPool(t)
	ctx := context.Background()
	repo := database.NewMembershipRepository(pool)
	orgID, orgEmail, schoolID := mkSchool(t, pool, "mem-active")

	if _, err := repo.Create(ctx, schoolID, &orgID, orgEmail, "owner", "active"); err != nil {
		t.Fatalf("create membership: %v", err)
	}
	got, err := repo.ActiveForOrganizer(ctx, orgID)
	if err != nil {
		t.Fatalf("active for organizer: %v", err)
	}
	if got.SchoolID != schoolID || got.Role != "owner" || got.Status != "active" {
		t.Fatalf("unexpected membership: %+v", got)
	}
	if got.OrganizerID == nil || *got.OrganizerID != orgID {
		t.Fatalf("organizer_id mismatch: %v", got.OrganizerID)
	}
}

func TestMembershipActiveForOrganizerNotFound(t *testing.T) {
	pool := tenantTestPool(t)
	_, err := database.NewMembershipRepository(pool).ActiveForOrganizer(context.Background(), "00000000-0000-0000-0000-0000000009fe")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// The invite → first-login flow: an admin adds an employee by email (a pending
// membership with no organizer), and on that person's first sign-in we link it.
func TestMembershipLinkPendingByEmail(t *testing.T) {
	pool := tenantTestPool(t)
	ctx := context.Background()
	repo := database.NewMembershipRepository(pool)
	orgID, orgEmail, schoolID := mkSchool(t, pool, "mem-link")

	if _, err := repo.Create(ctx, schoolID, nil, orgEmail, "member", "pending"); err != nil {
		t.Fatalf("create pending: %v", err)
	}
	// Before linking, the organizer has no active membership.
	if _, err := repo.ActiveForOrganizer(ctx, orgID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected no active membership pre-link, got %v", err)
	}

	if err := repo.LinkPendingByEmail(ctx, orgID, orgEmail); err != nil {
		t.Fatalf("link pending: %v", err)
	}
	got, err := repo.ActiveForOrganizer(ctx, orgID)
	if err != nil {
		t.Fatalf("active after link: %v", err)
	}
	if got.Email != orgEmail || got.Status != "active" || got.OrganizerID == nil || *got.OrganizerID != orgID {
		t.Fatalf("link didn't activate: %+v", got)
	}
}

func TestMembershipListBySchool(t *testing.T) {
	pool := tenantTestPool(t)
	ctx := context.Background()
	repo := database.NewMembershipRepository(pool)
	orgID, orgEmail, schoolID := mkSchool(t, pool, "mem-list")

	if _, err := repo.Create(ctx, schoolID, &orgID, orgEmail, "owner", "active"); err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if _, err := repo.Create(ctx, schoolID, nil, "invitee-list@tenant.test", "member", "pending"); err != nil {
		t.Fatalf("create invitee: %v", err)
	}
	list, err := repo.ListBySchool(ctx, schoolID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 members, got %d", len(list))
	}
}
