// Package challenge is the usecase for teacher-authored challenges (core + custom)
// and the student browse/list surfaces. School-scoped, mirroring usecase/slot.
package challenge

import (
	"context"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.ChallengeRepository
}

func New(repo repository.ChallengeRepository) *UseCase {
	return &UseCase{repo: repo}
}

// validate enforces the type invariants shared by Create and Update: a custom
// challenge is assigned to one student and carries no badge; a core challenge
// carries a badge and has no assignee. Title is always required. These are the
// same invariants the repository re-checks at the write boundary.
func validate(c model.Challenge) error {
	if c.Title == "" {
		return apperrors.NewValidation("title required")
	}
	switch c.Type {
	case "core":
		if c.BadgeLabel == "" {
			return apperrors.NewValidation("core challenge requires badge_label")
		}
		if c.AssignedDJID != "" {
			return apperrors.NewValidation("core challenge cannot have an assignee")
		}
	case "custom":
		if c.AssignedDJID == "" {
			return apperrors.NewValidation("custom challenge requires assigned_dj_id")
		}
		if c.BadgeLabel != "" || c.BadgeIcon != "" {
			return apperrors.NewValidation("custom challenge cannot carry a badge")
		}
	default:
		return apperrors.NewValidation(`type must be "core" or "custom"`)
	}
	return nil
}

// List returns every non-archived challenge for the school (teacher view).
func (uc *UseCase) List(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return uc.repo.List(ctx, schoolID)
}

func (uc *UseCase) Get(ctx context.Context, id, schoolID string) (model.Challenge, error) {
	return uc.repo.Get(ctx, id, schoolID)
}

func (uc *UseCase) Create(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	if err := validate(c); err != nil {
		return model.Challenge{}, err
	}
	return uc.repo.Create(ctx, c, schoolID)
}

func (uc *UseCase) Update(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	if err := validate(c); err != nil {
		return model.Challenge{}, err
	}
	return uc.repo.Update(ctx, c, schoolID)
}

func (uc *UseCase) Archive(ctx context.Context, id, schoolID string) error {
	return uc.repo.Archive(ctx, id, schoolID)
}

// ListCore returns the school's core (badge-bearing) challenges (Browse Core).
func (uc *UseCase) ListCore(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return uc.repo.ListCore(ctx, schoolID)
}

// ListForStudent returns the challenges relevant to one student: the school's
// core set followed by the custom challenges assigned to them. The two repo
// queries are composed here so the student sees one merged list.
func (uc *UseCase) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Challenge, error) {
	core, err := uc.repo.ListCore(ctx, schoolID)
	if err != nil {
		return nil, err
	}
	custom, err := uc.repo.ListForStudent(ctx, djID, schoolID)
	if err != nil {
		return nil, err
	}
	return append(core, custom...), nil
}
