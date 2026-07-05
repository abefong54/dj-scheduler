// Package submission is the usecase for student attempts: post (with host-
// allowlist media resolution), teacher review (pass/needs_work + badge snapshot),
// the school feed, and a student's own list. School-scoped, mirroring
// usecase/slot.
package submission

import (
	"context"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.SubmissionRepository
}

func New(repo repository.SubmissionRepository) *UseCase {
	return &UseCase{repo: repo}
}

// Post resolves the pasted media link (host allowlist) and writes the one active
// submission for (djID, challengeID) as 'posted'. A re-post overwrites the prior
// attempt. An unsupported/invalid link returns a validation error (→ 400).
func (uc *UseCase) Post(ctx context.Context, djID, schoolID, challengeID, mediaURL, caption string) (model.Submission, error) {
	if challengeID == "" {
		return model.Submission{}, apperrors.NewValidation("challenge_id is required")
	}
	media, err := ResolveMedia(mediaURL)
	if err != nil {
		return model.Submission{}, err
	}
	s := model.Submission{
		ChallengeID:   challengeID,
		DJID:          djID,
		MediaKind:     "embed",
		MediaProvider: media.Provider,
		MediaRef:      media.Ref,
		Caption:       caption,
	}
	return uc.repo.UpsertForChallenge(ctx, s, schoolID)
}

// Review records a teacher pass/needs_work. The badge snapshot for a core pass is
// handled in the repository (SQL). Status must be one of the two valid values.
func (uc *UseCase) Review(ctx context.Context, id, schoolID, status, note string) (model.Submission, error) {
	if status != "passed" && status != "needs_work" {
		return model.Submission{}, apperrors.NewValidation(`status must be "passed" or "needs_work"`)
	}
	return uc.repo.Review(ctx, id, schoolID, status, note, "", "")
}

// Feed returns the school's passed submissions, newest first.
func (uc *UseCase) Feed(ctx context.Context, schoolID string) ([]model.Submission, error) {
	return uc.repo.ListFeed(ctx, schoolID)
}

// ListForStudent returns one student's submissions across all their challenges.
func (uc *UseCase) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Submission, error) {
	return uc.repo.ListForStudent(ctx, djID, schoolID)
}
