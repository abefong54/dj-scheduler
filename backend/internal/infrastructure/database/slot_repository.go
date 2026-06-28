package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type slotRepo struct{ pool *pgxpool.Pool }

func NewSlotRepository(pool *pgxpool.Pool) *slotRepo {
	return &slotRepo{pool: pool}
}

func (r *slotRepo) List(ctx context.Context, eventID string) ([]model.Slot, error) {
	rows, err := r.pool.Query(ctx, querySlotList, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := []model.Slot{}
	for rows.Next() {
		var s model.Slot
		if err := rows.Scan(
			&s.ID, &s.EventID, &s.StageID, &s.StageName,
			&s.DjID, &s.DjName, &s.Genre,
			&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes,
		); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (r *slotRepo) Get(ctx context.Context, id, eventID string) (model.Slot, error) {
	var s model.Slot
	err := r.pool.QueryRow(ctx, querySlotGet, id, eventID).Scan(
		&s.ID, &s.EventID, &s.StageID, &s.StageName,
		&s.DjID, &s.DjName, &s.Genre,
		&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Slot{}, apperrors.ErrNotFound
	}
	return s, err
}

func (r *slotRepo) Create(ctx context.Context, s model.Slot, eventID string) (model.Slot, error) {
	err := r.pool.QueryRow(ctx, querySlotInsert,
		eventID, s.StageID, s.DjID, s.Genre, s.SlotDate, s.StartTime, s.EndTime, s.Notes).
		Scan(&s.ID)
	if err != nil {
		return model.Slot{}, err
	}
	s.EventID = eventID
	return s, nil
}

func (r *slotRepo) Update(ctx context.Context, s model.Slot, eventID string) (model.Slot, error) {
	err := r.pool.QueryRow(ctx, querySlotUpdate,
		s.StageID, s.DjID, s.Genre, s.SlotDate, s.StartTime, s.EndTime, s.Notes,
		s.ID, eventID,
	).Scan(
		&s.ID, &s.EventID, &s.StageID, &s.StageName,
		&s.DjID, &s.DjName, &s.Genre,
		&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Slot{}, apperrors.ErrNotFound
	}
	if err != nil {
		return model.Slot{}, err
	}
	return s, nil
}

func (r *slotRepo) Delete(ctx context.Context, id, eventID string) error {
	tag, err := r.pool.Exec(ctx, querySlotDelete, id, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
