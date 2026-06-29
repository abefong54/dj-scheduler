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

func (r *eventRepo) List(ctx context.Context, organizerID string) ([]model.Event, error) {
	rows, err := r.pool.Query(ctx, queryEventList, organizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []model.Event{}
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate, &e.Genres, &e.LineNotifyEnabled); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *eventRepo) Get(ctx context.Context, id, organizerID string) (model.Event, error) {
	var e model.Event
	err := r.pool.QueryRow(ctx, queryEventGet, id, organizerID).
		Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate, &e.Genres, &e.LineNotifyEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Event{}, apperrors.ErrNotFound
	}
	return e, err
}

func (r *eventRepo) GetPublic(ctx context.Context, id string) (model.Event, error) {
	var e model.Event
	err := r.pool.QueryRow(ctx, queryEventGetPublic, id).
		Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate, &e.Genres)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Event{}, apperrors.ErrNotFound
	}
	return e, err
}

func (r *eventRepo) Create(ctx context.Context, e model.Event, organizerID string) (model.Event, error) {
	var out model.Event
	err := r.pool.QueryRow(ctx, queryEventInsert,
		e.Name, e.VenueName, e.StartDate, e.EndDate, e.Genres, organizerID).
		Scan(&out.ID, &out.Name, &out.VenueName, &out.StartDate, &out.EndDate, &out.Genres, &out.LineNotifyEnabled)
	return out, err
}

func (r *eventRepo) Update(ctx context.Context, e model.Event, organizerID string) (model.Event, error) {
	var out model.Event
	err := r.pool.QueryRow(ctx, queryEventUpdate,
		e.Name, e.VenueName, e.StartDate, e.EndDate, e.Genres, e.ID, organizerID).
		Scan(&out.ID, &out.Name, &out.VenueName, &out.StartDate, &out.EndDate, &out.Genres, &out.LineNotifyEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Event{}, apperrors.ErrNotFound
	}
	return out, err
}

// SetLineNotifyToken stores or clears (encToken == nil) the encrypted LINE
// Notify token for the organizer's event. See US-006.
func (r *eventRepo) SetLineNotifyToken(ctx context.Context, id, organizerID string, encToken *string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, querySetLineToken, encToken, id, organizerID).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, apperrors.ErrNotFound
	}
	return enabled, err
}

func (r *eventRepo) Delete(ctx context.Context, id, organizerID string) error {
	_, err := r.pool.Exec(ctx, queryEventDelete, id, organizerID)
	return err
}

// Clone creates a new event from origID as a template: it copies the name
// (prefixed with "Copy of "), venue, genres, and stage structure, resets the
// dates to today, and intentionally does NOT copy slots (US-008).
func (r *eventRepo) Clone(ctx context.Context, origID, organizerID string) (model.Event, error) {
	var orig model.Event
	err := r.pool.QueryRow(ctx, queryEventCloneFetch, origID, organizerID).
		Scan(&orig.Name, &orig.VenueName, &orig.Genres)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Event{}, apperrors.ErrNotFound
	}
	if err != nil {
		return model.Event{}, err
	}

	var newEvent model.Event
	err = r.pool.QueryRow(ctx, queryEventCloneInsert,
		orig.Name, orig.VenueName, orig.Genres, organizerID).
		Scan(&newEvent.ID, &newEvent.Name, &newEvent.VenueName, &newEvent.StartDate, &newEvent.EndDate, &newEvent.Genres, &newEvent.LineNotifyEnabled)
	if err != nil {
		return model.Event{}, err
	}

	rows, err := r.pool.Query(ctx, queryStageListForClone, origID)
	if err != nil {
		return newEvent, err
	}
	defer rows.Close()
	for rows.Next() {
		var s model.Stage
		if scanErr := rows.Scan(&s.Name, &s.Color, &s.DisplayOrder); scanErr != nil {
			log.Printf("clone stage scan: %v", scanErr)
			continue
		}
		if _, execErr := r.pool.Exec(ctx, queryStageInsertForClone,
			newEvent.ID, s.Name, s.Color, s.DisplayOrder); execErr != nil {
			log.Printf("clone stage insert: %v", execErr)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("clone stage rows error: %v", err)
	}

	return newEvent, nil
}
