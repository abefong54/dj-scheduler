package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type challengeRepo struct{ pool *pgxpool.Pool }

func NewChallengeRepository(pool *pgxpool.Pool) *challengeRepo {
	return &challengeRepo{pool: pool}
}

// scanChallenge reads one challenge row in the challengeCols column order.
func scanChallenge(row pgx.Row) (model.Challenge, error) {
	var c model.Challenge
	err := row.Scan(
		&c.ID, &c.SchoolID, &c.Type, &c.Title, &c.Description,
		&c.Skill, &c.AssignedDJID, &c.BadgeLabel, &c.BadgeIcon,
		&c.SortOrder, &c.CreatedBy, &c.Archived, &c.CreatedAt,
	)
	return c, err
}

// validateChallengeInvariants enforces the type invariants at the write boundary
// (the usecase enforces them too): a custom challenge is assigned to one student
// and carries no badge; a core challenge carries a badge and has no assignee. It
// is a defense-in-depth guard so a bad row can never reach the table.
func validateChallengeInvariants(c model.Challenge) error {
	switch c.Type {
	case "core":
		if c.BadgeLabel == "" {
			return apperrors.NewValidation("core challenge requires badge_label")
		}
		if c.AssignedDJID != "" {
			return apperrors.NewValidation("core challenge cannot have an assignee")
		}
	case "custom":
		if c.AssignedDJID == "" {
			return apperrors.NewValidation("custom challenge requires assigned_dj_id")
		}
		if c.BadgeLabel != "" || c.BadgeIcon != "" {
			return apperrors.NewValidation("custom challenge cannot carry a badge")
		}
	default:
		return apperrors.NewValidation(`type must be "core" or "custom"`)
	}
	return nil
}

func (r *challengeRepo) queryList(ctx context.Context, sql string, args ...any) ([]model.Challenge, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Challenge{}
	for rows.Next() {
		c, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// List returns every non-archived challenge (core + custom) for the school.
func (r *challengeRepo) List(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return r.queryList(ctx, queryChallengeList, schoolID)
}

func (r *challengeRepo) Get(ctx context.Context, id, schoolID string) (model.Challenge, error) {
	c, err := scanChallenge(r.pool.QueryRow(ctx, queryChallengeGet, id, schoolID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Challenge{}, apperrors.ErrNotFound
	}
	return c, err
}

func (r *challengeRepo) Create(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	if err := validateChallengeInvariants(c); err != nil {
		return model.Challenge{}, err
	}
	c.SchoolID = schoolID
	err := r.pool.QueryRow(ctx, queryChallengeInsert,
		schoolID, c.Type, c.Title, c.Description, c.Skill,
		c.AssignedDJID, c.BadgeLabel, c.BadgeIcon, c.SortOrder, c.CreatedBy,
	).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return model.Challenge{}, err
	}
	return c, nil
}

func (r *challengeRepo) Update(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	if err := validateChallengeInvariants(c); err != nil {
		return model.Challenge{}, err
	}
	updated, err := scanChallenge(r.pool.QueryRow(ctx, queryChallengeUpdate,
		c.Type, c.Title, c.Description, c.Skill,
		c.AssignedDJID, c.BadgeLabel, c.BadgeIcon, c.SortOrder,
		c.ID, schoolID,
	))
	// No row updated means the challenge isn't this school's — report not-found so
	// cross-school access can't distinguish "missing" from "someone else's".
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Challenge{}, apperrors.ErrNotFound
	}
	return updated, err
}

// Archive soft-deletes a challenge (archived=true); earned badges survive because
// they are snapshotted onto submissions, not read from the live challenge.
func (r *challengeRepo) Archive(ctx context.Context, id, schoolID string) error {
	tag, err := r.pool.Exec(ctx, queryChallengeArchive, id, schoolID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// ListCore returns the school's core (badge-bearing) challenges for "Browse Core".
func (r *challengeRepo) ListCore(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return r.queryList(ctx, queryChallengeListCore, schoolID)
}

// ListForStudent returns the student's own open custom challenges (assigned to
// them, non-archived, school-scoped). The usecase merges these with the core set.
func (r *challengeRepo) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Challenge, error) {
	return r.queryList(ctx, queryChallengeListForStudent, djID, schoolID)
}
