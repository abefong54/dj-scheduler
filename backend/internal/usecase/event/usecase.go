package event

import (
	"context"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.EventRepository
}

func New(repo repository.EventRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, organizerID string) ([]model.Event, error) {
	return uc.repo.List(ctx, organizerID)
}

func (uc *UseCase) Get(ctx context.Context, id, organizerID string) (model.Event, error) {
	return uc.repo.Get(ctx, id, organizerID)
}

// GetPublic reads an event without organizer scoping, for the public schedule.
func (uc *UseCase) GetPublic(ctx context.Context, id string) (model.Event, error) {
	return uc.repo.GetPublic(ctx, id)
}

func (uc *UseCase) Create(ctx context.Context, e model.Event, organizerID string) (model.Event, error) {
	return uc.repo.Create(ctx, e, organizerID)
}

func (uc *UseCase) Update(ctx context.Context, e model.Event, organizerID string) (model.Event, error) {
	return uc.repo.Update(ctx, e, organizerID)
}

func (uc *UseCase) Delete(ctx context.Context, id, organizerID string) error {
	return uc.repo.Delete(ctx, id, organizerID)
}

func (uc *UseCase) Duplicate(ctx context.Context, id, organizerID string) (model.Event, error) {
	return uc.repo.Duplicate(ctx, id, organizerID)
}
