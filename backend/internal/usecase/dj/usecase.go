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

func (uc *UseCase) List(ctx context.Context) ([]model.DJ, error) {
	return uc.repo.List(ctx)
}

func (uc *UseCase) Get(ctx context.Context, id string) (model.DJ, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *UseCase) Create(ctx context.Context, name string, tags []string) (model.DJ, error) {
	return uc.repo.Create(ctx, name, tags)
}

func (uc *UseCase) Update(ctx context.Context, dj model.DJ) (model.DJ, error) {
	return uc.repo.Update(ctx, dj)
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
