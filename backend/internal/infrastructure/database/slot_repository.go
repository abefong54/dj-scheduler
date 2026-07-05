package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type slotRepo struct{ pool *pgxpool.Pool }

func NewSlotRepository(pool *pgxpool.Pool) *slotRepo {
	return &slotRepo{pool: pool}
}

// overlapConflict translates a Postgres exclusion-constraint violation (raised by
// the slots_stage_no_overlap / slots_dj_no_overlap GiST constraints) into a
// ConflictError, so a slot that loses the check-then-write race still surfaces as
// a 409 rather than a 500. Other errors pass through unchanged.
func overlapConflict(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
		typ := "stage_overlap"
		if pgErr.ConstraintName == "slots_dj_no_overlap" {
			typ = "dj_double_booked"
		}
		return &apperrors.ConflictError{Type: typ}
	}
	return err
}

// maxWriteRetries bounds how many times a slot write is retried after Postgres
// aborts it with a deadlock (40P01). When two overlapping slots are inserted at
// once they race on the GiST exclusion constraint; Postgres usually lets one win
// and rejects the loser with a clean 23P01 exclusion violation, but it can
// instead break the tie by killing one side as a deadlock victim. A deadlock is
// transient — on retry the winner has already committed, so the retry
// deterministically hits 23P01 and overlapConflict maps it to a 409. A small
// bound is plenty: once a winner commits there is nothing left to deadlock on.
const maxWriteRetries = 5

// isDeadlock reports whether err is a Postgres deadlock (SQLSTATE 40P01).
func isDeadlock(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40P01"
}

// retryOnDeadlock runs write, retrying only when Postgres aborts it as a deadlock
// victim. Any other outcome — success or a non-deadlock error — is returned
// immediately. The last deadlock error is returned if the retries are exhausted.
func retryOnDeadlock(write func() error) error {
	var err error
	for attempt := 0; attempt < maxWriteRetries; attempt++ {
		if err = write(); !isDeadlock(err) {
			return err
		}
	}
	return err
}

func (r *slotRepo) List(ctx context.Context, eventID, schoolID string) ([]model.Slot, error) {
	rows, err := r.pool.Query(ctx, querySlotList, eventID, schoolID)
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
			&s.DJConfirmation,
		); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// An empty result is ambiguous: the event may be the organizer's but have no
	// slots, or it may belong to someone else. Surface the latter as not-found.
	if len(slots) == 0 {
		owned, err := eventOwnedBy(ctx, r.pool, eventID, schoolID)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, apperrors.ErrNotFound
		}
	}
	return slots, nil
}

// ListPublic returns an event's slots without organizer scoping, for the
// unauthenticated public schedule. Callers must gate access some other way.
func (r *slotRepo) ListPublic(ctx context.Context, eventID string) ([]model.Slot, error) {
	rows, err := r.pool.Query(ctx, querySlotPublicList, eventID)
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
			&s.DJConfirmation,
		); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (r *slotRepo) Get(ctx context.Context, id, eventID, schoolID string) (model.Slot, error) {
	var s model.Slot
	err := r.pool.QueryRow(ctx, querySlotGet, id, eventID, schoolID).Scan(
		&s.ID, &s.EventID, &s.StageID, &s.StageName,
		&s.DjID, &s.DjName, &s.Genre,
		&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes,
		&s.DJConfirmation,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Slot{}, apperrors.ErrNotFound
	}
	return s, err
}

// GetPublicByID looks up a single slot by its id without organizer/event
// scoping, for the unauthenticated per-DJ share card (EL-049). It returns
// apperrors.ErrNotFound when no slot matches.
func (r *slotRepo) GetPublicByID(ctx context.Context, id string) (model.Slot, error) {
	var s model.Slot
	err := r.pool.QueryRow(ctx, querySlotPublicGet, id).Scan(
		&s.ID, &s.EventID, &s.StageID, &s.StageName,
		&s.DjID, &s.DjName, &s.Genre,
		&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes,
		&s.DJConfirmation,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Slot{}, apperrors.ErrNotFound
	}
	return s, err
}

func (r *slotRepo) Create(ctx context.Context, s model.Slot, eventID, schoolID string) (model.Slot, error) {
	err := retryOnDeadlock(func() error {
		return r.pool.QueryRow(ctx, querySlotInsert,
			eventID, s.StageID, s.DjID, s.Genre, s.SlotDate, s.StartTime, s.EndTime, s.Notes, schoolID).
			Scan(&s.ID)
	})
	// No row inserted means the event isn't this organizer's — report not-found.
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Slot{}, apperrors.ErrNotFound
	}
	if err != nil {
		return model.Slot{}, overlapConflict(err)
	}
	s.EventID = eventID
	return s, nil
}

func (r *slotRepo) Update(ctx context.Context, s model.Slot, eventID, schoolID string) (model.Slot, error) {
	err := retryOnDeadlock(func() error {
		return r.pool.QueryRow(ctx, querySlotUpdate,
			s.StageID, s.DjID, s.Genre, s.SlotDate, s.StartTime, s.EndTime, s.Notes,
			s.ID, eventID, schoolID,
		).Scan(
			&s.ID, &s.EventID, &s.StageID, &s.StageName,
			&s.DjID, &s.DjName, &s.Genre,
			&s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes,
			&s.DJConfirmation,
		)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Slot{}, apperrors.ErrNotFound
	}
	if err != nil {
		return model.Slot{}, overlapConflict(err)
	}
	return s, nil
}

// SetDJConfirmation records a DJ's confirm/flag response on a slot, but only if
// the slot actually belongs to that DJ. Returns ErrForbidden when no row matches
// (the slot isn't theirs or doesn't exist), so the portal can't touch others' slots.
func (r *slotRepo) SetDJConfirmation(ctx context.Context, slotID, djID, confirmation string) error {
	tag, err := r.pool.Exec(ctx, querySlotSetDJConfirmation, confirmation, slotID, djID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrForbidden
	}
	return nil
}

func (r *slotRepo) Delete(ctx context.Context, id, eventID, schoolID string) error {
	tag, err := r.pool.Exec(ctx, querySlotDelete, id, eventID, schoolID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
