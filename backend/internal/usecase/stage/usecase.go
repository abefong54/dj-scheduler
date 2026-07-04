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

func (uc *UseCase) List(ctx context.Context, eventID, schoolID string) ([]model.Stage, error) {
	return uc.repo.List(ctx, eventID, schoolID)
}

// ListPublic returns stages without organizer scoping, for the public schedule.
func (uc *UseCase) ListPublic(ctx context.Context, eventID string) ([]model.Stage, error) {
	return uc.repo.ListPublic(ctx, eventID)
}

func (uc *UseCase) Get(ctx context.Context, id, eventID, schoolID string) (model.Stage, error) {
	return uc.repo.Get(ctx, id, eventID, schoolID)
}

func (uc *UseCase) Create(ctx context.Context, eventID, name, color, schoolID string) (model.Stage, error) {
	return uc.repo.Create(ctx, eventID, name, color, schoolID)
}

func (uc *UseCase) Update(ctx context.Context, s model.Stage, schoolID string) (model.Stage, error) {
	return uc.repo.Update(ctx, s, schoolID)
}

func (uc *UseCase) Delete(ctx context.Context, id, eventID, schoolID string) error {
	return uc.repo.Delete(ctx, id, eventID, schoolID)
}
