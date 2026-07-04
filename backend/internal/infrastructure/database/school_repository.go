package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type schoolRepo struct{ pool *pgxpool.Pool }

func NewSchoolRepository(pool *pgxpool.Pool) *schoolRepo {
	return &schoolRepo{pool: pool}
}

func (r *schoolRepo) Create(ctx context.Context, name, city, createdBy string) (model.School, error) {
	var s model.School
	err := r.pool.QueryRow(ctx, querySchoolInsert, name, city, createdBy).
		Scan(&s.ID, &s.Name, &s.City, &s.CreatedBy, &s.CreatedAt)
	return s, err
}

func (r *schoolRepo) Get(ctx context.Context, id string) (model.School, error) {
	var s model.School
	err := r.pool.QueryRow(ctx, querySchoolGet, id).
		Scan(&s.ID, &s.Name, &s.City, &s.CreatedBy, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.School{}, apperrors.ErrNotFound
	}
	return s, err
}
