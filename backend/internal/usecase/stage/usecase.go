package stage

import (
	"context"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.StageRepository
}

func New(repo repository.StageRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, eventID string) ([]model.Stage, error) {
	return uc.repo.List(ctx, eventID)
}

func (uc *UseCase) Get(ctx context.Context, id, eventID string) (model.Stage, error) {
	return uc.repo.Get(ctx, id, eventID)
}

func (uc *UseCase) Create(ctx context.Context, eventID, name, color string) (model.Stage, error) {
	return uc.repo.Create(ctx, eventID, name, color)
}

func (uc *UseCase) Update(ctx context.Context, s model.Stage) (model.Stage, error) {
	return uc.repo.Update(ctx, s)
}

func (uc *UseCase) Delete(ctx context.Context, id, eventID string) error {
	return uc.repo.Delete(ctx, id, eventID)
}
