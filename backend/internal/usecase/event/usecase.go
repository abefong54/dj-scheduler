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

func (uc *UseCase) List(ctx context.Context, schoolID string) ([]model.Event, error) {
	return uc.repo.List(ctx, schoolID)
}

func (uc *UseCase) Get(ctx context.Context, id, schoolID string) (model.Event, error) {
	return uc.repo.Get(ctx, id, schoolID)
}

// GetPublic reads an event without school scoping, for the public schedule.
func (uc *UseCase) GetPublic(ctx context.Context, id string) (model.Event, error) {
	return uc.repo.GetPublic(ctx, id)
}

func (uc *UseCase) Create(ctx context.Context, e model.Event, createdBy, schoolID string) (model.Event, error) {
	return uc.repo.Create(ctx, e, createdBy, schoolID)
}

func (uc *UseCase) Update(ctx context.Context, e model.Event, schoolID string) (model.Event, error) {
	return uc.repo.Update(ctx, e, schoolID)
}

func (uc *UseCase) Delete(ctx context.Context, id, schoolID string) error {
	return uc.repo.Delete(ctx, id, schoolID)
}

func (uc *UseCase) Clone(ctx context.Context, id, createdBy, schoolID string) (model.Event, error) {
	return uc.repo.Clone(ctx, id, createdBy, schoolID)
}
