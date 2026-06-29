package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type djRepo struct{ pool *pgxpool.Pool }

func NewDJRepository(pool *pgxpool.Pool) *djRepo {
	return &djRepo{pool: pool}
}

func (r *djRepo) List(ctx context.Context, organizerID string) ([]model.DJ, error) {
	rows, err := r.pool.Query(ctx, queryDJList, organizerID)
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

func (r *djRepo) Get(ctx context.Context, id, organizerID string) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJGet, id, organizerID).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.DJ{}, apperrors.ErrNotFound
	}
	return d, err
}

func (r *djRepo) Create(ctx context.Context, name string, tags []string, organizerID string) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJInsert, name, tags, organizerID).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	return d, err
}

func (r *djRepo) Update(ctx context.Context, dj model.DJ, organizerID string) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJUpdate, dj.Name, dj.GenreTags, dj.ID, organizerID).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.DJ{}, apperrors.ErrNotFound
	}
	return d, err
}

func (r *djRepo) Delete(ctx context.Context, id, organizerID string) error {
	_, err := r.pool.Exec(ctx, queryDJDelete, id, organizerID)
	return err
}

// SetPortalToken stores a portal token hash + expiry for a DJ the organizer owns.
// Returns ErrNotFound if the DJ does not exist or belongs to another organizer.
func (r *djRepo) SetPortalToken(ctx context.Context, djID, organizerID, tokenHash string, expiresAt time.Time) error {
	tag, err := r.pool.Exec(ctx, queryDJSetPortalToken, tokenHash, expiresAt, djID, organizerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// GetByPortalToken returns the DJ whose (non-expired) token hash matches, or
// ErrNotFound if there is no match. It is intentionally NOT organizer-scoped:
// the token itself is the credential for the public portal.
func (r *djRepo) GetByPortalToken(ctx context.Context, tokenHash string) (model.DJ, error) {
	var d model.DJ
	err := r.pool.QueryRow(ctx, queryDJGetByPortalToken, tokenHash).
		Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.DJ{}, apperrors.ErrNotFound
	}
	return d, err
}

// ListSlotsForDJ returns every slot booked for a DJ across all events, with the
// event and stage names resolved, ordered by date then start time.
func (r *djRepo) ListSlotsForDJ(ctx context.Context, djID string) ([]model.PortalSlot, error) {
	rows, err := r.pool.Query(ctx, queryDJPortalSlots, djID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	slots := []model.PortalSlot{}
	for rows.Next() {
		var s model.PortalSlot
		if err := rows.Scan(&s.ID, &s.EventID, &s.EventName, &s.StageName, &s.Genre,
			&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes, &s.DJConfirmation); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}
