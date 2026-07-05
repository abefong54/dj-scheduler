package studentauth

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
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

// stubGoogle stands in for the Google OAuth dance: UserFromCode always returns a
// fixed profile (mirrors usecase/auth/usecase_test.go).
type stubGoogle struct{ user GoogleUser }

func (s stubGoogle) AuthCodeURL(string) string { return "" }
func (s stubGoogle) UserFromCode(context.Context, string) (GoogleUser, error) {
	return s.user, nil
}

// seedSchool creates a throwaway organizer + school and returns the school id,
// registering FK-safe cleanup.
func seedSchool(t *testing.T, pool *pgxpool.Pool, key string) string {
	t.Helper()
	ctx := context.Background()
	org, err := database.NewOrganizerRepository(pool).Create(ctx, key+"@studentauth.test", "Owner "+key, key)
	if err != nil {
		t.Fatalf("seed organizer: %v", err)
	}
	school, err := database.NewSchoolRepository(pool).Create(ctx, "Student Auth School "+key, "", org.ID)
	if err != nil {
		t.Fatalf("seed school: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM djs WHERE school_id = $1", school.ID)
		pool.Exec(ctx, "DELETE FROM schools WHERE id = $1", school.ID)
		pool.Exec(ctx, "DELETE FROM organizers WHERE id = $1", org.ID)
	})
	return school.ID
}

// TestHandleCallbackResolution exercises the three resolution outcomes of the
// student login path: a returning student (matched by google_id), a first-time
// student (matched by roster email, google_id auto-linked), and a Google user on
// no roster (rejected with ErrNotFound → 403). DB-backed so the real SQL is
// covered; skips cleanly without DATABASE_URL.
func TestHandleCallbackResolution(t *testing.T) {
	pool := dbPool(t)
	ctx := context.Background()
	djRepo := database.NewDJRepository(pool)
	schoolID := seedSchool(t, pool, "resolution")

	const secret = "studentauth-test-secret-studentauth-test-01"

	tests := []struct {
		name         string
		email        string // roster email set by the teacher ("" = student not on roster)
		bindGoogleID string // pre-bound google_id (returning student); "" = first login
		guGoogleID   string // google_id the OAuth callback presents
		guEmail      string // email the OAuth callback presents
		wantNotFound bool
		wantLinked   bool // after callback, guGoogleID should resolve to the seeded DJ
	}{
		{
			name:         "returning student matched by google_id",
			email:        "returning@studentauth.test",
			bindGoogleID: "returning-gid",
			guGoogleID:   "returning-gid",
			guEmail:      "returning@studentauth.test",
			wantLinked:   true,
		},
		{
			name:       "first login auto-links google_id by email",
			email:      "firstlogin@studentauth.test",
			guGoogleID: "firstlogin-gid",
			guEmail:    "firstlogin@studentauth.test",
			wantLinked: true,
		},
		{
			name:         "google user on no roster is rejected",
			guGoogleID:   "stranger-gid",
			guEmail:      "stranger@studentauth.test",
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Seed the roster DJ (with email, and google_id for a returning student).
			if tt.email != "" {
				dj, err := djRepo.Create(ctx, "Test Student", nil, schoolID)
				if err != nil {
					t.Fatalf("create dj: %v", err)
				}
				if err := djRepo.SetEmail(ctx, dj.ID, schoolID, tt.email); err != nil {
					t.Fatalf("set email: %v", err)
				}
				if tt.bindGoogleID != "" {
					if err := djRepo.SetGoogleID(ctx, dj.ID, tt.bindGoogleID); err != nil {
						t.Fatalf("set google_id: %v", err)
					}
				}
			}

			uc := New(
				stubGoogle{user: GoogleUser{GoogleID: tt.guGoogleID, Email: tt.guEmail, Name: "Test Student"}},
				djRepo, secret, time.Hour,
			)

			jwt, err := uc.HandleCallback(ctx, "code")

			if tt.wantNotFound {
				if !errors.Is(err, apperrors.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got err=%v jwt=%q", err, jwt)
				}
				return
			}
			if err != nil {
				t.Fatalf("HandleCallback: %v", err)
			}
			if jwt == "" {
				t.Fatal("expected a signed JWT, got empty string")
			}
			if tt.wantLinked {
				linked, err := djRepo.FindByGoogleID(ctx, tt.guGoogleID)
				if err != nil {
					t.Fatalf("expected google_id bound after callback: %v", err)
				}
				if linked.SchoolID != schoolID {
					t.Fatalf("expected linked DJ scoped to school %s, got %s", schoolID, linked.SchoolID)
				}
			}
		})
	}
}
