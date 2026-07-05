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

func (uc *UseCase) List(ctx context.Context, schoolID string, filter repository.DJListFilter) ([]model.DJ, error) {
	return uc.repo.List(ctx, schoolID, filter)
}

func (uc *UseCase) Get(ctx context.Context, id, schoolID string) (model.DJ, error) {
	return uc.repo.Get(ctx, id, schoolID)
}

func (uc *UseCase) Create(ctx context.Context, name string, tags []string, schoolID string) (model.DJ, error) {
	return uc.repo.Create(ctx, name, tags, schoolID)
}

func (uc *UseCase) Update(ctx context.Context, dj model.DJ, schoolID string) (model.DJ, error) {
	return uc.repo.Update(ctx, dj, schoolID)
}

func (uc *UseCase) Delete(ctx context.Context, id, schoolID string) error {
	return uc.repo.Delete(ctx, id, schoolID)
}
