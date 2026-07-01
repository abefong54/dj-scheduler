package slot

import (
	"context"
	"errors"
	"testing"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

// fakeRepo is a full SlotRepository whose List returns a fixed set and whose
// writes record whether they were called.
type fakeRepo struct {
	stored        []model.Slot
	created       bool
	updated       bool
	createErr     error
	getErr        error
	publicSlot    model.Slot
	publicByIDErr error
}

func (f *fakeRepo) List(_ context.Context, _, _ string) ([]model.Slot, error) {
	return f.stored, nil
}
func (f *fakeRepo) ListPublic(_ context.Context, _ string) ([]model.Slot, error) {
	return f.stored, nil
}
func (f *fakeRepo) Get(_ context.Context, _, _, _ string) (model.Slot, error) {
	return model.Slot{}, f.getErr
}
func (f *fakeRepo) GetPublicByID(_ context.Context, _ string) (model.Slot, error) {
	return f.publicSlot, f.publicByIDErr
}
func (f *fakeRepo) Create(_ context.Context, s model.Slot, _, _ string) (model.Slot, error) {
	f.created = true
	return s, f.createErr
}
func (f *fakeRepo) Update(_ context.Context, s model.Slot, _, _ string) (model.Slot, error) {
	f.updated = true
	return s, nil
}
func (f *fakeRepo) Delete(_ context.Context, _, _, _ string) error            { return nil }
func (f *fakeRepo) SetDJConfirmation(_ context.Context, _, _, _ string) error { return nil }

func bookedSlot() model.Slot {
	return model.Slot{
		ID: "slot-a", EventID: "evt-1", StageID: "stage-a", DjID: "dj-sasha",
		SlotDate: "2026-08-15", StartTime: "22:00", EndTime: "23:00",
	}
}

func TestGetPublicByID_ReturnsSlot(t *testing.T) {
	want := bookedSlot()
	repo := &fakeRepo{publicSlot: want}
	uc := New(repo)

	got, err := uc.GetPublicByID(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.ID != want.ID || got.EventID != want.EventID {
		t.Fatalf("expected slot %+v, got %+v", want, got)
	}
}

func TestGetPublicByID_NotFound(t *testing.T) {
	repo := &fakeRepo{publicByIDErr: apperrors.ErrNotFound}
	uc := New(repo)

	_, err := uc.GetPublicByID(context.Background(), "missing")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreate_RejectsConflictWithoutWriting(t *testing.T) {
	repo := &fakeRepo{stored: []model.Slot{bookedSlot()}}
	uc := New(repo)

	_, err := uc.Create(context.Background(), model.Slot{
		StageID: "stage-b", DjID: "dj-sasha", SlotDate: "2026-08-15",
		StartTime: "22:30", EndTime: "23:30",
	}, "evt-1", "")

	if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if repo.created {
		t.Error("repo.Create was called despite a conflict")
	}
}

func TestCreate_AllowsNonConflicting(t *testing.T) {
	repo := &fakeRepo{stored: []model.Slot{bookedSlot()}}
	uc := New(repo)

	_, err := uc.Create(context.Background(), model.Slot{
		StageID: "stage-a", DjID: "dj-sasha", SlotDate: "2026-08-15",
		StartTime: "23:00", EndTime: "23:30", // adjacent
	}, "evt-1", "")

	if err != nil {
		t.Fatalf("non-conflicting create should succeed, got %v", err)
	}
	if !repo.created {
		t.Error("repo.Create was not called for a valid slot")
	}
}

func TestUpdate_ExcludesSelfFromConflict(t *testing.T) {
	repo := &fakeRepo{stored: []model.Slot{bookedSlot()}}
	uc := New(repo)

	_, err := uc.Update(context.Background(), model.Slot{
		ID: "slot-a", StageID: "stage-a", DjID: "dj-sasha", SlotDate: "2026-08-15",
		StartTime: "22:00", EndTime: "23:00", // unchanged
	}, "evt-1", "")

	if err != nil {
		t.Fatalf("updating a slot to its own time should succeed, got %v", err)
	}
	if !repo.updated {
		t.Error("repo.Update was not called")
	}
}

func TestUpdate_MissingSlotReturnsNotFoundNotConflict(t *testing.T) {
	// The slot being updated does not exist, but its payload would overlap an
	// existing slot. Not-found must win over conflict.
	repo := &fakeRepo{
		stored: []model.Slot{bookedSlot()},
		getErr: apperrors.ErrNotFound,
	}
	uc := New(repo)

	_, err := uc.Update(context.Background(), model.Slot{
		ID: "missing", StageID: "stage-a", DjID: "dj-sasha", SlotDate: "2026-08-15",
		StartTime: "22:30", EndTime: "23:30",
	}, "evt-1", "")

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if repo.updated {
		t.Error("repo.Update was called for a missing slot")
	}
}

func TestUpdate_RejectsConflictWithoutWriting(t *testing.T) {
	repo := &fakeRepo{stored: []model.Slot{bookedSlot(), {
		ID: "slot-b", EventID: "evt-1", StageID: "stage-a", DjID: "dj-other",
		SlotDate: "2026-08-15", StartTime: "20:00", EndTime: "21:00",
	}}}
	uc := New(repo)

	// Move slot-b onto stage-a at a time that overlaps slot-a.
	_, err := uc.Update(context.Background(), model.Slot{
		ID: "slot-b", StageID: "stage-a", DjID: "dj-other", SlotDate: "2026-08-15",
		StartTime: "22:30", EndTime: "23:30",
	}, "evt-1", "")

	if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if repo.updated {
		t.Error("repo.Update was called despite a conflict")
	}
}
