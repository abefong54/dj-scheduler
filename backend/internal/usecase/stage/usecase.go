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

func (uc *UseCase) List(ctx context.Context, eventID, organizerID string) ([]model.Stage, error) {
	return uc.repo.List(ctx, eventID, organizerID)
}

// ListPublic returns stages without organizer scoping, for the public schedule.
func (uc *UseCase) ListPublic(ctx context.Context, eventID string) ([]model.Stage, error) {
	return uc.repo.ListPublic(ctx, eventID)
}

func (uc *UseCase) Get(ctx context.Context, id, eventID, organizerID string) (model.Stage, error) {
	return uc.repo.Get(ctx, id, eventID, organizerID)
}

func (uc *UseCase) Create(ctx context.Context, eventID, name, color, organizerID string) (model.Stage, error) {
	return uc.repo.Create(ctx, eventID, name, color, organizerID)
}

func (uc *UseCase) Update(ctx context.Context, s model.Stage, organizerID string) (model.Stage, error) {
	return uc.repo.Update(ctx, s, organizerID)
}

func (uc *UseCase) Delete(ctx context.Context, id, eventID, organizerID string) error {
	return uc.repo.Delete(ctx, id, eventID, organizerID)
}
