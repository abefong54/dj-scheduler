// Package submission is the usecase for student attempts: post (with host-
// allowlist media resolution), teacher review (pass/needs_work + badge snapshot),
// the school feed, and a student's own list. School-scoped, mirroring
// usecase/slot. SCAFFOLD: shapes only — Track C (submission backend) fills the
// logic, including the media resolver.
package submission

import (
	"context"
	"errors"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

var errNotImplemented = errors.New("not implemented")

type UseCase struct {
	repo repository.SubmissionRepository
}

func New(repo repository.SubmissionRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Create(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error) {
	return model.Submission{}, errNotImplemented
}

func (uc *UseCase) Get(ctx context.Context, id, schoolID string) (model.Submission, error) {
	return model.Submission{}, errNotImplemented
}

// UpsertForChallenge posts or re-works the one active submission for a challenge.
func (uc *UseCase) UpsertForChallenge(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error) {
	return model.Submission{}, errNotImplemented
}

// Review records a teacher pass/needs_work; on a core pass the badge is snapshotted.
func (uc *UseCase) Review(ctx context.Context, id, schoolID, status, note, badgeLabel, badgeIcon string) (model.Submission, error) {
	return model.Submission{}, errNotImplemented
}

// ListFeed returns the school's passed submissions, newest first.
func (uc *UseCase) ListFeed(ctx context.Context, schoolID string) ([]model.Submission, error) {
	return nil, errNotImplemented
}

// ListForStudent returns one student's submissions across all their challenges.
func (uc *UseCase) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Submission, error) {
	return nil, errNotImplemented
}
