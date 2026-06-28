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
	s.EventID = eventID
	if err := CheckConflicts(ctx, uc.repo, s, ""); err != nil {
		return model.Slot{}, err
	}
	return uc.repo.Create(ctx, s, eventID)
}

func (uc *UseCase) Update(ctx context.Context, s model.Slot, eventID string) (model.Slot, error) {
	// Confirm the slot exists first so updating a missing slot returns not-found
	// rather than a conflict against some other slot its payload would overlap.
	if _, err := uc.repo.Get(ctx, s.ID, eventID); err != nil {
		return model.Slot{}, err
	}
	s.EventID = eventID
	if err := CheckConflicts(ctx, uc.repo, s, s.ID); err != nil {
		return model.Slot{}, err
	}
	return uc.repo.Update(ctx, s, eventID)
}

func (uc *UseCase) Delete(ctx context.Context, id, eventID string) error {
	return uc.repo.Delete(ctx, id, eventID)
}
