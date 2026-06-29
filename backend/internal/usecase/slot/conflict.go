package slot

import (
	"context"
	"strconv"
	"strings"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

// slotLister is the slice of the slot repository that conflict checking needs.
type slotLister interface {
	List(ctx context.Context, eventID, organizerID string) ([]model.Slot, error)
}

// CheckConflicts returns an *apperrors.ConflictError if candidate would
// double-book its DJ or overlap another slot on the same stage within the same
// event and date. excludeSlotID is the candidate's own ID on update (pass "" on
// create) so a slot never conflicts with itself.
//
// DJ double-booking is checked first to match the API contract, so a candidate
// that clashes both by DJ and by stage reports "dj_double_booked".
func CheckConflicts(ctx context.Context, repo slotLister, candidate model.Slot, excludeSlotID, organizerID string) error {
	stored, err := repo.List(ctx, candidate.EventID, organizerID)
	if err != nil {
		return err
	}

	// Pass 1: DJ double-booked on the same date, any stage.
	if candidate.DjID != "" {
		for _, s := range stored {
			if s.ID == excludeSlotID || s.SlotDate != candidate.SlotDate {
				continue
			}
			if s.DjID == candidate.DjID && overlaps(candidate, s) {
				return &apperrors.ConflictError{Type: "dj_double_booked", ConflictingSlotID: s.ID}
			}
		}
	}

	// Pass 2: overlap on the same stage, same date.
	for _, s := range stored {
		if s.ID == excludeSlotID || s.SlotDate != candidate.SlotDate {
			continue
		}
		if s.StageID == candidate.StageID && overlaps(candidate, s) {
			return &apperrors.ConflictError{Type: "stage_overlap", ConflictingSlotID: s.ID}
		}
	}

	return nil
}

// overlaps reports whether two slots' time ranges intersect with exclusive ends,
// so adjacent slots (one ending exactly when the next begins) do not conflict.
func overlaps(a, b model.Slot) bool {
	aStart, aEnd := minutes(a.StartTime), minutes(a.EndTime)
	bStart, bEnd := minutes(b.StartTime), minutes(b.EndTime)
	return aStart < bEnd && bStart < aEnd
}

// minutes converts an "HH:MM" or "HH:MM:SS" time string to minutes since
// midnight. Unparseable input yields -1, which simply never overlaps a real slot.
func minutes(t string) int {
	parts := strings.Split(t, ":")
	if len(parts) < 2 {
		return -1
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}
	return h*60 + m
}
