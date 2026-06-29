package slot

import (
	"context"
	"errors"
	"testing"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type fakeLister struct {
	slots []model.Slot
	err   error
}

func (f fakeLister) List(_ context.Context, _ string) ([]model.Slot, error) {
	return f.slots, f.err
}

// existing returns a stable set of stored slots for a single event/date.
func existing() []model.Slot {
	return []model.Slot{
		{
			ID:        "slot-a",
			EventID:   "evt-1",
			StageID:   "stage-a",
			DjID:      "dj-sasha",
			SlotDate:  "2026-08-15",
			StartTime: "22:00",
			EndTime:   "23:00",
		},
	}
}

func TestCheckConflicts_DJDoubleBookedAcrossStages(t *testing.T) {
	repo := fakeLister{slots: existing()}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-b", // different stage
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-15",
		StartTime: "22:30",
		EndTime:   "23:30",
	}

	err := CheckConflicts(context.Background(), repo, candidate, "")

	var conflict *apperrors.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected ConflictError, got %v", err)
	}
	if conflict.Type != "dj_double_booked" {
		t.Errorf("type = %q, want dj_double_booked", conflict.Type)
	}
	if conflict.ConflictingSlotID != "slot-a" {
		t.Errorf("conflicting_slot_id = %q, want slot-a", conflict.ConflictingSlotID)
	}
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("errors.Is(err, ErrConflict) = false, want true")
	}
}

func TestCheckConflicts_StageOverlap(t *testing.T) {
	repo := fakeLister{slots: existing()}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-a", // same stage
		DjID:      "dj-other",
		SlotDate:  "2026-08-15",
		StartTime: "22:30",
		EndTime:   "23:30",
	}

	err := CheckConflicts(context.Background(), repo, candidate, "")

	var conflict *apperrors.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected ConflictError, got %v", err)
	}
	if conflict.Type != "stage_overlap" {
		t.Errorf("type = %q, want stage_overlap", conflict.Type)
	}
	if conflict.ConflictingSlotID != "slot-a" {
		t.Errorf("conflicting_slot_id = %q, want slot-a", conflict.ConflictingSlotID)
	}
}

func TestCheckConflicts_AdjacentSlotsAllowed(t *testing.T) {
	repo := fakeLister{slots: existing()}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-a",
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-15",
		StartTime: "23:00", // starts exactly when slot-a ends
		EndTime:   "23:30",
	}

	if err := CheckConflicts(context.Background(), repo, candidate, ""); err != nil {
		t.Fatalf("adjacent slots should not conflict, got %v", err)
	}
}

func TestCheckConflicts_UpdateExcludesSelf(t *testing.T) {
	repo := fakeLister{slots: existing()}
	candidate := model.Slot{
		ID:        "slot-a",
		EventID:   "evt-1",
		StageID:   "stage-a",
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-15",
		StartTime: "22:00",
		EndTime:   "23:00",
	}

	if err := CheckConflicts(context.Background(), repo, candidate, "slot-a"); err != nil {
		t.Fatalf("slot should not conflict with itself, got %v", err)
	}
}

func TestCheckConflicts_DifferentDateNoConflict(t *testing.T) {
	repo := fakeLister{slots: existing()}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-a",
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-16", // next day
		StartTime: "22:30",
		EndTime:   "23:30",
	}

	if err := CheckConflicts(context.Background(), repo, candidate, ""); err != nil {
		t.Fatalf("different date should not conflict, got %v", err)
	}
}

func TestCheckConflicts_NonOverlappingTimesNoConflict(t *testing.T) {
	repo := fakeLister{slots: existing()}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-a",
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-15",
		StartTime: "20:00",
		EndTime:   "21:00", // ends before slot-a starts
	}

	if err := CheckConflicts(context.Background(), repo, candidate, ""); err != nil {
		t.Fatalf("non-overlapping times should not conflict, got %v", err)
	}
}

func TestCheckConflicts_EmptyDJIgnoredForDoubleBooking(t *testing.T) {
	// Two slots with no DJ assigned on different stages must not be treated as
	// a DJ double-booking just because both dj_id are empty.
	repo := fakeLister{slots: []model.Slot{{
		ID:        "slot-a",
		EventID:   "evt-1",
		StageID:   "stage-a",
		DjID:      "",
		SlotDate:  "2026-08-15",
		StartTime: "22:00",
		EndTime:   "23:00",
	}}}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-b",
		DjID:      "",
		SlotDate:  "2026-08-15",
		StartTime: "22:30",
		EndTime:   "23:30",
	}

	if err := CheckConflicts(context.Background(), repo, candidate, ""); err != nil {
		t.Fatalf("empty DJ on different stages should not conflict, got %v", err)
	}
}

func TestCheckConflicts_StageOverlapAcrossMidnight(t *testing.T) {
	// A set running 22:30 -> 00:30 (past midnight) overlaps a 23:00–23:45 slot
	// on the same stage and date.
	repo := fakeLister{slots: []model.Slot{{
		ID:        "late",
		EventID:   "evt-1",
		StageID:   "stage-a",
		SlotDate:  "2026-08-15",
		StartTime: "23:00",
		EndTime:   "23:45",
	}}}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-a",
		SlotDate:  "2026-08-15",
		StartTime: "22:30",
		EndTime:   "00:30", // wraps into the next day
	}

	err := CheckConflicts(context.Background(), repo, candidate, "")
	var conflict *apperrors.ConflictError
	if !errors.As(err, &conflict) || conflict.Type != "stage_overlap" {
		t.Fatalf("expected stage_overlap across midnight, got %v", err)
	}
}

func TestCheckConflicts_DJDoubleBookedAcrossDateBoundary(t *testing.T) {
	// A set that runs 23:30 (Aug 15) -> 00:30 (Aug 16) collides in real time with
	// another booking for the same DJ at 00:00–01:00 on Aug 16.
	repo := fakeLister{slots: []model.Slot{{
		ID:        "night",
		EventID:   "evt-1",
		StageID:   "stage-a",
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-15",
		StartTime: "23:30",
		EndTime:   "00:30",
	}}}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-b",
		DjID:      "dj-sasha",
		SlotDate:  "2026-08-16",
		StartTime: "00:00",
		EndTime:   "01:00",
	}

	err := CheckConflicts(context.Background(), repo, candidate, "")
	var conflict *apperrors.ConflictError
	if !errors.As(err, &conflict) || conflict.Type != "dj_double_booked" {
		t.Fatalf("expected dj_double_booked across the date boundary, got %v", err)
	}
}

func TestCheckConflicts_WrappedSlotsThatDoNotOverlap(t *testing.T) {
	// An early-evening slot and a genuinely separate after-midnight set on the
	// same stage/date must not be reported as a conflict.
	repo := fakeLister{slots: []model.Slot{{
		ID:        "early",
		EventID:   "evt-1",
		StageID:   "stage-a",
		SlotDate:  "2026-08-15",
		StartTime: "18:00",
		EndTime:   "19:00",
	}}}
	candidate := model.Slot{
		EventID:   "evt-1",
		StageID:   "stage-a",
		SlotDate:  "2026-08-15",
		StartTime: "23:30",
		EndTime:   "00:30",
	}

	if err := CheckConflicts(context.Background(), repo, candidate, ""); err != nil {
		t.Fatalf("non-overlapping wrapped slot should not conflict, got %v", err)
	}
}
