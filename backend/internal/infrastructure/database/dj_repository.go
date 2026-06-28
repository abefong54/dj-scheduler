package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type djRepo struct{ pool *pgxpool.Pool }

func NewDJRepository(pool *pgxpool.Pool) *djRepo {
	return &djRepo{pool: pool}
}

func (r *djRepo) List(ctx context.Context) ([]model.DJ, error) {
	rows, err := r.pool.Query(ctx, queryDJList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	djs := []model.DJ{}
	for rows.Next() {
		var d model.DJ
		if err := rows.Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt); err != nil {
			return nil, err
		}
		djs = append(djs, d)
	}
	return djs, rows.Err()
}

func (r *djRepo) Get(ctx context.Context, id string) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJGet, id).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.DJ{}, apperrors.ErrNotFound
	}
	return d, err
}

func (r *djRepo) Create(ctx context.Context, name string, tags []string) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJInsert, name, tags).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	return d, err
}

func (r *djRepo) Update(ctx context.Context, dj model.DJ) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJUpdate, dj.Name, dj.GenreTags, dj.ID).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.DJ{}, apperrors.ErrNotFound
	}
	return d, err
}

func (r *djRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, queryDJDelete, id)
	return err
}
