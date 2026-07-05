// Package challenge is the usecase for teacher-authored challenges (core + custom)
// and the student browse/list surfaces. School-scoped, mirroring usecase/slot.
// SCAFFOLD: shapes only — Track B (challenge backend) fills the logic.
package challenge

import (
	"context"
	"errors"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

var errNotImplemented = errors.New("not implemented")

type UseCase struct {
	repo repository.ChallengeRepository
}

func New(repo repository.ChallengeRepository) *UseCase {
	return &UseCase{repo: repo}
}

// List returns every non-archived challenge for the school (teacher view).
func (uc *UseCase) List(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return nil, errNotImplemented
}

func (uc *UseCase) Get(ctx context.Context, id, schoolID string) (model.Challenge, error) {
	return model.Challenge{}, errNotImplemented
}

func (uc *UseCase) Create(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	return model.Challenge{}, errNotImplemented
}

func (uc *UseCase) Update(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	return model.Challenge{}, errNotImplemented
}

func (uc *UseCase) Archive(ctx context.Context, id, schoolID string) error {
	return errNotImplemented
}

// ListCore returns the school's core (badge-bearing) challenges (Browse Core).
func (uc *UseCase) ListCore(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return nil, errNotImplemented
}

// ListForStudent returns the core set plus the student's assigned custom challenges.
func (uc *UseCase) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Challenge, error) {
	return nil, errNotImplemented
}
