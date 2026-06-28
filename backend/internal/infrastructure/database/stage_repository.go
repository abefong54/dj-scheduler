package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type stageRepo struct{ pool *pgxpool.Pool }

func NewStageRepository(pool *pgxpool.Pool) *stageRepo {
	return &stageRepo{pool: pool}
}

func (r *stageRepo) List(ctx context.Context, eventID string) ([]model.Stage, error) {
	rows, err := r.pool.Query(ctx, queryStageList, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stages := []model.Stage{}
	for rows.Next() {
		var s model.Stage
		if err := rows.Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder); err != nil {
			return nil, err
		}
		stages = append(stages, s)
	}
	return stages, rows.Err()
}

func (r *stageRepo) Get(ctx context.Context, id, eventID string) (model.Stage, error) {
	var s model.Stage
	err := r.pool.QueryRow(ctx, queryStageGet, id, eventID).
		Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Stage{}, apperrors.ErrNotFound
	}
	return s, err
}

func (r *stageRepo) Create(ctx context.Context, eventID, name, color string) (model.Stage, error) {
	var s model.Stage
	err := r.pool.QueryRow(ctx, queryStageInsert, eventID, name, color).
		Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
	return s, err
}

func (r *stageRepo) Update(ctx context.Context, s model.Stage) (model.Stage, error) {
	var out model.Stage
	err := r.pool.QueryRow(ctx, queryStageUpdate, s.Name, s.Color, s.ID, s.EventID).
		Scan(&out.ID, &out.EventID, &out.Name, &out.Color, &out.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Stage{}, apperrors.ErrNotFound
	}
	return out, err
}

func (r *stageRepo) Delete(ctx context.Context, id, eventID string) error {
	_, err := r.pool.Exec(ctx, queryStageDelete, id, eventID)
	return err
}
