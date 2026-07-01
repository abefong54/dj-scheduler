// Package performance aggregates played slots into per-student stage-time reps,
// so a teacher can see who is getting performance time and who is under-served
// (EL-043, Wedge B). All reads are scoped to the organizer by the repository.
package performance

import (
	"context"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

// DefaultUnderservedThreshold is the rep count below which a student counts as
// under-served when the caller doesn't supply one.
const DefaultUnderservedThreshold = 3

type UseCase struct {
	repo repository.PerformanceRepository
}

func New(repo repository.PerformanceRepository) *UseCase {
	return &UseCase{repo: repo}
}

// DJPerformance returns one DJ's reps across the organizer's events.
func (uc *UseCase) DJPerformance(ctx context.Context, djID, organizerID string) (model.DJPerformance, error) {
	return uc.repo.DJPerformance(ctx, djID, organizerID)
}

// RosterSummary returns every active student's reps in the window.
func (uc *UseCase) RosterSummary(ctx context.Context, organizerID string, filter repository.PerformanceFilter) ([]model.RosterPerformance, error) {
	return uc.repo.RosterSummary(ctx, organizerID, filter)
}

// Underserved returns the organizer's active students whose rep count in the
// window is strictly below threshold. A threshold <= 0 falls back to the default.
// Graduates are already excluded by RosterSummary (is_student = true only).
func (uc *UseCase) Underserved(ctx context.Context, organizerID string, filter repository.PerformanceFilter, threshold int) ([]model.RosterPerformance, error) {
	if threshold <= 0 {
		threshold = DefaultUnderservedThreshold
	}
	roster, err := uc.repo.RosterSummary(ctx, organizerID, filter)
	if err != nil {
		return nil, err
	}
	out := []model.RosterPerformance{}
	for _, r := range roster {
		if r.Reps < threshold {
			out = append(out, r)
		}
	}
	return out, nil
}
