package slot

import (
	"context"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.SlotRepository
}

func New(repo repository.SlotRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, eventID string) ([]model.Slot, error) {
	return uc.repo.List(ctx, eventID)
}

func (uc *UseCase) Get(ctx context.Context, id, eventID string) (model.Slot, error) {
	return uc.repo.Get(ctx, id, eventID)
}

func (uc *UseCase) Create(ctx context.Context, s model.Slot, eventID string) (model.Slot, error) {
	return uc.repo.Create(ctx, s, eventID)
}

func (uc *UseCase) Update(ctx context.Context, s model.Slot, eventID string) (model.Slot, error) {
	return uc.repo.Update(ctx, s, eventID)
}

func (uc *UseCase) Delete(ctx context.Context, id, eventID string) error {
	return uc.repo.Delete(ctx, id, eventID)
}
