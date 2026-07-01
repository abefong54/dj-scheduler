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

func (uc *UseCase) List(ctx context.Context, eventID, organizerID string) ([]model.Slot, error) {
	return uc.repo.List(ctx, eventID, organizerID)
}

// ListPublic returns slots without organizer scoping, for the public schedule.
func (uc *UseCase) ListPublic(ctx context.Context, eventID string) ([]model.Slot, error) {
	return uc.repo.ListPublic(ctx, eventID)
}

func (uc *UseCase) Get(ctx context.Context, id, eventID, organizerID string) (model.Slot, error) {
	return uc.repo.Get(ctx, id, eventID, organizerID)
}

// GetPublicByID reads a slot by id alone, without organizer scoping, for the
// public per-DJ share card (EL-049).
func (uc *UseCase) GetPublicByID(ctx context.Context, id string) (model.Slot, error) {
	return uc.repo.GetPublicByID(ctx, id)
}

func (uc *UseCase) Create(ctx context.Context, s model.Slot, eventID, organizerID string) (model.Slot, error) {
	s.EventID = eventID
	if err := CheckConflicts(ctx, uc.repo, s, "", organizerID); err != nil {
		return model.Slot{}, err
	}
	return uc.repo.Create(ctx, s, eventID, organizerID)
}

func (uc *UseCase) Update(ctx context.Context, s model.Slot, eventID, organizerID string) (model.Slot, error) {
	// Confirm the slot exists first so updating a missing slot returns not-found
	// rather than a conflict against some other slot its payload would overlap.
	// Get is organizer-scoped, so another organizer's slot also reads as not-found.
	if _, err := uc.repo.Get(ctx, s.ID, eventID, organizerID); err != nil {
		return model.Slot{}, err
	}
	s.EventID = eventID
	if err := CheckConflicts(ctx, uc.repo, s, s.ID, organizerID); err != nil {
		return model.Slot{}, err
	}
	return uc.repo.Update(ctx, s, eventID, organizerID)
}

func (uc *UseCase) Delete(ctx context.Context, id, eventID, organizerID string) error {
	return uc.repo.Delete(ctx, id, eventID, organizerID)
}

// SetDJConfirmation records a DJ's confirm/flag response on their own slot
// (US-011). djID must be the DJ resolved from a valid portal token; the
// repository enforces that the slot belongs to that DJ.
func (uc *UseCase) SetDJConfirmation(ctx context.Context, slotID, djID, confirmation string) error {
	return uc.repo.SetDJConfirmation(ctx, slotID, djID, confirmation)
}
