package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type organizerRepo struct{ pool *pgxpool.Pool }

func NewOrganizerRepository(pool *pgxpool.Pool) *organizerRepo {
	return &organizerRepo{pool: pool}
}

// scanOrganizer reads the standard organizer column list (id, email, name,
// google_id, line_id, created_at). The nullable provider ids are COALESCE'd to
// ” in SQL so they scan into the string model.
func scanOrganizer(row pgx.Row) (model.Organizer, error) {
	var o model.Organizer
	err := row.Scan(&o.ID, &o.Email, &o.Name, &o.GoogleID, &o.LineID, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organizer{}, apperrors.ErrNotFound
	}
	return o, err
}

func (r *organizerRepo) FindByGoogleID(ctx context.Context, googleID string) (model.Organizer, error) {
	return scanOrganizer(r.pool.QueryRow(ctx, queryOrganizerFindByGoogleID, googleID))
}

func (r *organizerRepo) FindByLineID(ctx context.Context, lineID string) (model.Organizer, error) {
	return scanOrganizer(r.pool.QueryRow(ctx, queryOrganizerFindByLineID, lineID))
}

func (r *organizerRepo) FindByEmail(ctx context.Context, email string) (model.Organizer, error) {
	return scanOrganizer(r.pool.QueryRow(ctx, queryOrganizerFindByEmail, email))
}

func (r *organizerRepo) FindByID(ctx context.Context, id string) (model.Organizer, error) {
	return scanOrganizer(r.pool.QueryRow(ctx, queryOrganizerFindByID, id))
}

func (r *organizerRepo) Create(ctx context.Context, email, name, googleID string) (model.Organizer, error) {
	return scanOrganizer(r.pool.QueryRow(ctx, queryOrganizerInsert, email, name, googleID))
}

// CreateWithLineID creates a LINE-first organizer (google_id NULL).
func (r *organizerRepo) CreateWithLineID(ctx context.Context, email, name, lineID string) (model.Organizer, error) {
	return scanOrganizer(r.pool.QueryRow(ctx, queryOrganizerInsertLine, email, name, lineID))
}

// LinkLineID attaches a LINE identity to an existing organizer (the same-email
// account-merge case).
func (r *organizerRepo) LinkLineID(ctx context.Context, organizerID, lineID string) error {
	_, err := r.pool.Exec(ctx, queryOrganizerLinkLineID, organizerID, lineID)
	return err
}
