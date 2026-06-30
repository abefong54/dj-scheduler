package performance_test

import (
	"context"
	"errors"
	"testing"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
	perf "eventlineup/internal/usecase/performance"
)

// fakeRepo is an in-memory PerformanceRepository for unit-testing the usecase
// logic (threshold/default) without a database.
type fakeRepo struct {
	roster    []model.RosterPerformance
	djPerf    model.DJPerformance
	err       error
	gotFilter repository.PerformanceFilter
}

func (f *fakeRepo) DJPerformance(_ context.Context, _, _ string) (model.DJPerformance, error) {
	return f.djPerf, f.err
}

func (f *fakeRepo) RosterSummary(_ context.Context, _ string, filter repository.PerformanceFilter) ([]model.RosterPerformance, error) {
	f.gotFilter = filter
	return f.roster, f.err
}

func ids(rows []model.RosterPerformance) []string {
	out := []string{}
	for _, r := range rows {
		out = append(out, r.DjID)
	}
	return out
}

func TestUnderservedKeepsOnlyBelowThreshold(t *testing.T) {
	repo := &fakeRepo{roster: []model.RosterPerformance{
		{DjID: "a", IsStudent: true, Reps: 0},
		{DjID: "b", IsStudent: true, Reps: 2},
		{DjID: "c", IsStudent: true, Reps: 3}, // exactly threshold → NOT under-served
		{DjID: "d", IsStudent: true, Reps: 9},
	}}
	got, err := perf.New(repo).Underserved(context.Background(), "org1", repository.PerformanceFilter{}, 3)
	if err != nil {
		t.Fatalf("Underserved: %v", err)
	}
	if want := []string{"a", "b"}; !equal(ids(got), want) {
		t.Fatalf("threshold 3: got %v, want %v", ids(got), want)
	}
}

func TestUnderservedDefaultsThresholdWhenNonPositive(t *testing.T) {
	repo := &fakeRepo{roster: []model.RosterPerformance{
		{DjID: "a", Reps: 2}, // below default 3
		{DjID: "b", Reps: 3}, // at default → kept out
		{DjID: "c", Reps: 8},
	}}
	for _, threshold := range []int{0, -1} {
		got, err := perf.New(repo).Underserved(context.Background(), "org1", repository.PerformanceFilter{}, threshold)
		if err != nil {
			t.Fatalf("Underserved(%d): %v", threshold, err)
		}
		if want := []string{"a"}; !equal(ids(got), want) {
			t.Fatalf("threshold %d should default to %d: got %v, want %v",
				threshold, perf.DefaultUnderservedThreshold, ids(got), want)
		}
	}
}

func TestUnderservedPassesFilterThrough(t *testing.T) {
	repo := &fakeRepo{}
	f := repository.PerformanceFilter{EventID: "e1", From: "2026-01-01", To: "2026-12-31"}
	if _, err := perf.New(repo).Underserved(context.Background(), "org1", f, 3); err != nil {
		t.Fatalf("Underserved: %v", err)
	}
	if repo.gotFilter != f {
		t.Fatalf("filter not forwarded to repo: got %+v, want %+v", repo.gotFilter, f)
	}
}

func TestUnderservedPropagatesRepoError(t *testing.T) {
	sentinel := errors.New("boom")
	repo := &fakeRepo{err: sentinel}
	if _, err := perf.New(repo).Underserved(context.Background(), "org1", repository.PerformanceFilter{}, 3); !errors.Is(err, sentinel) {
		t.Fatalf("want repo error propagated, got %v", err)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
