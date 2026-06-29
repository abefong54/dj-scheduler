package dj

import (
	"context"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.DJRepository
}

func New(repo repository.DJRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, organizerID string, filter repository.DJListFilter) ([]model.DJ, error) {
	return uc.repo.List(ctx, organizerID, filter)
}

func (uc *UseCase) Get(ctx context.Context, id, organizerID string) (model.DJ, error) {
	return uc.repo.Get(ctx, id, organizerID)
}

func (uc *UseCase) Create(ctx context.Context, name string, tags []string, organizerID string) (model.DJ, error) {
	return uc.repo.Create(ctx, name, tags, organizerID)
}

func (uc *UseCase) Update(ctx context.Context, dj model.DJ, organizerID string) (model.DJ, error) {
	return uc.repo.Update(ctx, dj, organizerID)
}

func (uc *UseCase) Delete(ctx context.Context, id, organizerID string) error {
	return uc.repo.Delete(ctx, id, organizerID)
}
