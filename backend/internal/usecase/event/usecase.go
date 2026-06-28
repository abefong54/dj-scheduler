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

func (uc *UseCase) List(ctx context.Context) ([]model.Event, error) {
	return uc.repo.List(ctx)
}

func (uc *UseCase) Get(ctx context.Context, id string) (model.Event, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *UseCase) Create(ctx context.Context, e model.Event) (model.Event, error) {
	return uc.repo.Create(ctx, e)
}

func (uc *UseCase) Update(ctx context.Context, e model.Event) (model.Event, error) {
	return uc.repo.Update(ctx, e)
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *UseCase) Duplicate(ctx context.Context, id string) (model.Event, error) {
	return uc.repo.Duplicate(ctx, id)
}
