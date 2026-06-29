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

func (f fakeLister) List(_ context.Context, _, _ string) ([]model.Slot, error) {
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

	err := CheckConflicts(context.Background(), repo, candidate, "", "")

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

	err := CheckConflicts(context.Background(), repo, candidate, "", "")

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

	if err := CheckConflicts(context.Background(), repo, candidate, "", ""); err != nil {
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

	if err := CheckConflicts(context.Background(), repo, candidate, "slot-a", ""); err != nil {
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

	if err := CheckConflicts(context.Background(), repo, candidate, "", ""); err != nil {
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

	if err := CheckConflicts(context.Background(), repo, candidate, "", ""); err != nil {
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

	if err := CheckConflicts(context.Background(), repo, candidate, "", ""); err != nil {
		t.Fatalf("empty DJ on different stages should not conflict, got %v", err)
	}
}
