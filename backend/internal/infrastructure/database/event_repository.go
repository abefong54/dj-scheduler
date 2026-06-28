package database

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type eventRepo struct{ pool *pgxpool.Pool }

func NewEventRepository(pool *pgxpool.Pool) *eventRepo {
	return &eventRepo{pool: pool}
}

func (r *eventRepo) List(ctx context.Context) ([]model.Event, error) {
	rows, err := r.pool.Query(ctx, queryEventList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []model.Event{}
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate, &e.Genres); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *eventRepo) Get(ctx context.Context, id string) (model.Event, error) {
	var e model.Event
	err := r.pool.QueryRow(ctx, queryEventGet, id).
		Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate, &e.Genres)
	return e, err
}

func (r *eventRepo) Create(ctx context.Context, e model.Event) (model.Event, error) {
	var out model.Event
	err := r.pool.QueryRow(ctx, queryEventInsert,
		e.Name, e.VenueName, e.StartDate, e.EndDate, e.Genres).
		Scan(&out.ID, &out.Name, &out.VenueName, &out.StartDate, &out.EndDate, &out.Genres)
	return out, err
}

func (r *eventRepo) Update(ctx context.Context, e model.Event) (model.Event, error) {
	var out model.Event
	err := r.pool.QueryRow(ctx, queryEventUpdate,
		e.Name, e.VenueName, e.StartDate, e.EndDate, e.Genres, e.ID).
		Scan(&out.ID, &out.Name, &out.VenueName, &out.StartDate, &out.EndDate, &out.Genres)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Event{}, apperrors.ErrNotFound
	}
	return out, err
}

func (r *eventRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, queryEventDelete, id)
	return err
}

func (r *eventRepo) Duplicate(ctx context.Context, origID string) (model.Event, error) {
	var orig model.Event
	err := r.pool.QueryRow(ctx, queryEventDuplicateFetch, origID).
		Scan(&orig.Name, &orig.VenueName, &orig.StartDate, &orig.EndDate, &orig.Genres)
	if err != nil {
		return model.Event{}, err
	}

	var newEvent model.Event
	err = r.pool.QueryRow(ctx, queryEventDuplicateInsert,
		orig.Name, orig.VenueName, orig.StartDate, orig.EndDate, orig.Genres).
		Scan(&newEvent.ID, &newEvent.Name, &newEvent.VenueName, &newEvent.StartDate, &newEvent.EndDate, &newEvent.Genres)
	if err != nil {
		return model.Event{}, err
	}

	rows, err := r.pool.Query(ctx, queryStageListForDuplicate, origID)
	if err != nil {
		return newEvent, err
	}
	defer rows.Close()
	for rows.Next() {
		var s model.Stage
		rows.Scan(&s.Name, &s.Color, &s.DisplayOrder)
		if _, execErr := r.pool.Exec(ctx, queryStageInsertForDuplicate,
			newEvent.ID, s.Name, s.Color, s.DisplayOrder); execErr != nil {
			log.Printf("duplicate stage insert: %v", execErr)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("duplicate stage rows error: %v", err)
	}

	return newEvent, nil
}
