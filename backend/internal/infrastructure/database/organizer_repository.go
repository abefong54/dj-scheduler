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

func (r *organizerRepo) FindByGoogleID(ctx context.Context, googleID string) (model.Organizer, error) {
	var o model.Organizer
	err := r.pool.QueryRow(ctx, queryOrganizerFindByGoogleID, googleID).
		Scan(&o.ID, &o.Email, &o.Name, &o.GoogleID, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organizer{}, apperrors.ErrNotFound
	}
	return o, err
}

func (r *organizerRepo) Create(ctx context.Context, email, name, googleID string) (model.Organizer, error) {
	var o model.Organizer
	err := r.pool.QueryRow(ctx, queryOrganizerInsert, email, name, googleID).
		Scan(&o.ID, &o.Email, &o.Name, &o.GoogleID, &o.CreatedAt)
	return o, err
}
