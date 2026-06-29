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

func (r *stageRepo) List(ctx context.Context, eventID, organizerID string) ([]model.Stage, error) {
	rows, err := r.pool.Query(ctx, queryStageList, eventID, organizerID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// An empty result is ambiguous: the event may be the organizer's but have no
	// stages, or it may belong to someone else. Surface the latter as not-found.
	if len(stages) == 0 {
		owned, err := eventOwnedBy(ctx, r.pool, eventID, organizerID)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, apperrors.ErrNotFound
		}
	}
	return stages, nil
}

// ListPublic returns an event's stages without organizer scoping, for the
// unauthenticated public schedule. Callers must gate access some other way.
func (r *stageRepo) ListPublic(ctx context.Context, eventID string) ([]model.Stage, error) {
	rows, err := r.pool.Query(ctx, queryStagePublicList, eventID)
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

func (r *stageRepo) Get(ctx context.Context, id, eventID, organizerID string) (model.Stage, error) {
	var s model.Stage
	err := r.pool.QueryRow(ctx, queryStageGet, id, eventID, organizerID).
		Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Stage{}, apperrors.ErrNotFound
	}
	return s, err
}

func (r *stageRepo) Create(ctx context.Context, eventID, name, color, organizerID string) (model.Stage, error) {
	var s model.Stage
	err := r.pool.QueryRow(ctx, queryStageInsert, eventID, name, color, organizerID).
		Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
	// No row inserted means the event isn't this organizer's — report not-found
	// rather than leaking that the event exists.
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Stage{}, apperrors.ErrNotFound
	}
	return s, err
}

func (r *stageRepo) Update(ctx context.Context, s model.Stage, organizerID string) (model.Stage, error) {
	var out model.Stage
	err := r.pool.QueryRow(ctx, queryStageUpdate, s.Name, s.Color, s.ID, s.EventID, organizerID).
		Scan(&out.ID, &out.EventID, &out.Name, &out.Color, &out.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Stage{}, apperrors.ErrNotFound
	}
	return out, err
}

func (r *stageRepo) Delete(ctx context.Context, id, eventID, organizerID string) error {
	tag, err := r.pool.Exec(ctx, queryStageDelete, id, eventID, organizerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
