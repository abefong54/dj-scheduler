package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type submissionRepo struct{ pool *pgxpool.Pool }

func NewSubmissionRepository(pool *pgxpool.Pool) *submissionRepo {
	return &submissionRepo{pool: pool}
}

// scanSubmission reads one submission row in the submissionCols column order.
func scanSubmission(row pgx.Row) (model.Submission, error) {
	var s model.Submission
	err := row.Scan(
		&s.ID, &s.SchoolID, &s.ChallengeID, &s.DJID,
		&s.MediaKind, &s.MediaProvider, &s.MediaRef,
		&s.Caption, &s.Status, &s.ReviewNote,
		&s.EarnedBadgeLabel, &s.EarnedBadgeIcon,
		&s.ReviewedBy, &s.ReviewedAt, &s.CreatedAt,
	)
	return s, err
}

func (r *submissionRepo) queryList(ctx context.Context, sql string, args ...any) ([]model.Submission, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Submission{}
	for rows.Next() {
		s, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Create inserts a fresh submission (posted). Post-attempt uses UpsertForChallenge
// instead; Create exists for completeness/tests.
func (r *submissionRepo) Create(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error) {
	return scanSubmission(r.pool.QueryRow(ctx, querySubmissionUpsert,
		schoolID, s.ChallengeID, s.DJID, s.MediaKind, s.MediaProvider, s.MediaRef, s.Caption,
	))
}

func (r *submissionRepo) Get(ctx context.Context, id, schoolID string) (model.Submission, error) {
	s, err := scanSubmission(r.pool.QueryRow(ctx, querySubmissionGet, id, schoolID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Submission{}, apperrors.ErrNotFound
	}
	return s, err
}

// UpsertForChallenge writes the one active submission per (dj_id, challenge_id),
// resetting it to 'posted' on a re-post.
func (r *submissionRepo) UpsertForChallenge(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error) {
	return scanSubmission(r.pool.QueryRow(ctx, querySubmissionUpsert,
		schoolID, s.ChallengeID, s.DJID, s.MediaKind, s.MediaProvider, s.MediaRef, s.Caption,
	))
}

// Review records a pass/needs_work; the badge snapshot for a core pass happens in
// SQL (see querySubmissionReview). badgeLabel/badgeIcon are accepted for interface
// compatibility but the snapshot is computed from the challenge row, so they are
// unused. A missing row (wrong school) reads as not-found.
func (r *submissionRepo) Review(ctx context.Context, id, schoolID, status, note, badgeLabel, badgeIcon string) (model.Submission, error) {
	s, err := scanSubmission(r.pool.QueryRow(ctx, querySubmissionReview, id, schoolID, status, note))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Submission{}, apperrors.ErrNotFound
	}
	return s, err
}

// ListFeed returns the school's passed submissions, newest first.
func (r *submissionRepo) ListFeed(ctx context.Context, schoolID string) ([]model.Submission, error) {
	return r.queryList(ctx, querySubmissionListFeed, schoolID)
}

// ListForReview returns the school's posted submissions (awaiting sign-off),
// oldest first — the teacher review queue.
func (r *submissionRepo) ListForReview(ctx context.Context, schoolID string) ([]model.Submission, error) {
	return r.queryList(ctx, querySubmissionListForReview, schoolID)
}

// ListForStudent returns one student's submissions across all their challenges.
func (r *submissionRepo) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Submission, error) {
	return r.queryList(ctx, querySubmissionListForStudent, djID, schoolID)
}
