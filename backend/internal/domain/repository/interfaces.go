package repository

import (
	"context"
	"time"

	"eventlineup/internal/domain/model"
)

// DJListFilter narrows the DJ list (EL-019). The zero value lists everyone.
type DJListFilter struct {
	// CertifiedFor, when non-empty, keeps only DJs certified for that genre
	// (case-insensitive). ReadyOnly keeps only DJs with at least one certification.
	CertifiedFor string
	ReadyOnly    bool
}

type DJRepository interface {
	List(ctx context.Context, organizerID string, filter DJListFilter) ([]model.DJ, error)
	Get(ctx context.Context, id, organizerID string) (model.DJ, error)
	Create(ctx context.Context, name string, tags []string, organizerID string) (model.DJ, error)
	Update(ctx context.Context, dj model.DJ, organizerID string) (model.DJ, error)
	Delete(ctx context.Context, id, organizerID string) error

	// DJ portal tokens (US-009).
	SetPortalToken(ctx context.Context, djID, organizerID, tokenHash string, expiresAt time.Time) error
	GetByPortalToken(ctx context.Context, tokenHash string) (model.DJ, error)
	ListSlotsForDJ(ctx context.Context, djID string) ([]model.PortalSlot, error)
}

type EventRepository interface {
	List(ctx context.Context, organizerID string) ([]model.Event, error)
	Get(ctx context.Context, id, organizerID string) (model.Event, error)
	// GetPublic reads an event without organizer scoping, for the public schedule.
	GetPublic(ctx context.Context, id string) (model.Event, error)
	Create(ctx context.Context, e model.Event, organizerID string) (model.Event, error)
	Update(ctx context.Context, e model.Event, organizerID string) (model.Event, error)
	Delete(ctx context.Context, id, organizerID string) error
	Clone(ctx context.Context, id, organizerID string) (model.Event, error)
	// SetLineNotifyToken stores encToken (an already-encrypted token) for the
	// organizer's event, or clears it when encToken is nil. It returns whether
	// LINE Notify is now enabled, and apperrors.ErrNotFound if the event is not
	// the organizer's. See US-006.
	SetLineNotifyToken(ctx context.Context, id, organizerID string, encToken *string) (bool, error)
}

// StageRepository scopes every operation to the organizer who owns the parent
// event (EL-036); a stage that isn't theirs reads as not-found.
type StageRepository interface {
	List(ctx context.Context, eventID, organizerID string) ([]model.Stage, error)
	// ListPublic is unscoped, for the public schedule endpoint only.
	ListPublic(ctx context.Context, eventID string) ([]model.Stage, error)
	Get(ctx context.Context, id, eventID, organizerID string) (model.Stage, error)
	Create(ctx context.Context, eventID, name, color, organizerID string) (model.Stage, error)
	Update(ctx context.Context, s model.Stage, organizerID string) (model.Stage, error)
	Delete(ctx context.Context, id, eventID, organizerID string) error
}

// PerformanceFilter narrows a roster-wide performance summary to a window and/or
// a single event (EL-043). Empty fields mean "no bound".
type PerformanceFilter struct {
	EventID string // restrict to one event; "" = all events
	From    string // inclusive slot_date lower bound "YYYY-MM-DD"; "" = no lower bound
	To      string // inclusive slot_date upper bound; "" = no upper bound
}

// PerformanceRepository aggregates played slots into performance reps, always
// scoped to the organizer who owns the parent events (EL-036/EL-043).
type PerformanceRepository interface {
	// DJPerformance aggregates one DJ's reps across the organizer's events with a
	// by-genre breakdown. Returns apperrors.ErrNotFound if the DJ isn't the
	// organizer's; an owned DJ who never played returns a zero-rep summary.
	DJPerformance(ctx context.Context, djID, organizerID string) (model.DJPerformance, error)
	// RosterSummary returns every active student (is_student=true) of the
	// organizer with their rep count in the window, including zero-rep students.
	RosterSummary(ctx context.Context, organizerID string, filter PerformanceFilter) ([]model.RosterPerformance, error)
}

type LeadRepository interface {
	Create(ctx context.Context, lead model.Lead) (model.Lead, error)
}

type OrganizerRepository interface {
	FindByGoogleID(ctx context.Context, googleID string) (model.Organizer, error)
	Create(ctx context.Context, email, name, googleID string) (model.Organizer, error)
}

// SlotRepository scopes every operation to the organizer who owns the parent
// event (EL-036), except SetDJConfirmation which is scoped to the DJ instead.
type SlotRepository interface {
	List(ctx context.Context, eventID, organizerID string) ([]model.Slot, error)
	// ListPublic is unscoped, for the public schedule endpoint only.
	ListPublic(ctx context.Context, eventID string) ([]model.Slot, error)
	// GetPublicByID looks up a slot by id alone, without organizer/event scoping,
	// for the public per-DJ share card (EL-049).
	GetPublicByID(ctx context.Context, id string) (model.Slot, error)
	Get(ctx context.Context, id, eventID, organizerID string) (model.Slot, error)
	Create(ctx context.Context, s model.Slot, eventID, organizerID string) (model.Slot, error)
	Update(ctx context.Context, s model.Slot, eventID, organizerID string) (model.Slot, error)
	Delete(ctx context.Context, id, eventID, organizerID string) error
	// SetDJConfirmation records a DJ portal confirm/flag, scoped to the DJ who owns
	// the slot (ErrForbidden when the slot isn't theirs). US-011.
	SetDJConfirmation(ctx context.Context, slotID, djID, confirmation string) error
}
