package repository

import (
	"context"
	"time"

	"eventlineup/internal/domain/model"
)

type DJRepository interface {
	List(ctx context.Context, organizerID string) ([]model.DJ, error)
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
	Get(ctx context.Context, id, eventID, organizerID string) (model.Slot, error)
	Create(ctx context.Context, s model.Slot, eventID, organizerID string) (model.Slot, error)
	Update(ctx context.Context, s model.Slot, eventID, organizerID string) (model.Slot, error)
	Delete(ctx context.Context, id, eventID, organizerID string) error
	// SetDJConfirmation records a DJ portal confirm/flag, scoped to the DJ who owns
	// the slot (ErrForbidden when the slot isn't theirs). US-011.
	SetDJConfirmation(ctx context.Context, slotID, djID, confirmation string) error
}
