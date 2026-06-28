package repository

import (
	"context"

	"eventlineup/internal/domain/model"
)

type DJRepository interface {
	List(ctx context.Context, organizerID string) ([]model.DJ, error)
	Get(ctx context.Context, id, organizerID string) (model.DJ, error)
	Create(ctx context.Context, name string, tags []string, organizerID string) (model.DJ, error)
	Update(ctx context.Context, dj model.DJ, organizerID string) (model.DJ, error)
	Delete(ctx context.Context, id, organizerID string) error
}

type EventRepository interface {
	List(ctx context.Context, organizerID string) ([]model.Event, error)
	Get(ctx context.Context, id, organizerID string) (model.Event, error)
	// GetPublic reads an event without organizer scoping, for the public schedule.
	GetPublic(ctx context.Context, id string) (model.Event, error)
	Create(ctx context.Context, e model.Event, organizerID string) (model.Event, error)
	Update(ctx context.Context, e model.Event, organizerID string) (model.Event, error)
	Delete(ctx context.Context, id, organizerID string) error
	Duplicate(ctx context.Context, id, organizerID string) (model.Event, error)
}

type StageRepository interface {
	List(ctx context.Context, eventID string) ([]model.Stage, error)
	Get(ctx context.Context, id, eventID string) (model.Stage, error)
	Create(ctx context.Context, eventID, name, color string) (model.Stage, error)
	Update(ctx context.Context, s model.Stage) (model.Stage, error)
	Delete(ctx context.Context, id, eventID string) error
}

type OrganizerRepository interface {
	FindByGoogleID(ctx context.Context, googleID string) (model.Organizer, error)
	Create(ctx context.Context, email, name, googleID string) (model.Organizer, error)
}

type SlotRepository interface {
	List(ctx context.Context, eventID string) ([]model.Slot, error)
	Get(ctx context.Context, id, eventID string) (model.Slot, error)
	Create(ctx context.Context, s model.Slot, eventID string) (model.Slot, error)
	Update(ctx context.Context, s model.Slot, eventID string) (model.Slot, error)
	Delete(ctx context.Context, id, eventID string) error
}
