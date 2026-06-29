package slot

import (
	"context"
	"strconv"
	"strings"
	"time"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

const minutesPerDay = 24 * 60

// slotLister is the slice of the slot repository that conflict checking needs.
type slotLister interface {
	List(ctx context.Context, eventID, organizerID string) ([]model.Slot, error)
}

// CheckConflicts returns an *apperrors.ConflictError if candidate would
// double-book its DJ (any stage) or overlap another slot on the same stage
// within the event. excludeSlotID is the candidate's own ID on update (pass ""
// on create) so a slot never conflicts with itself.
//
// Slots are compared on an absolute timeline built from slot_date + time, so a
// set that runs past midnight (end <= start, meaning it ends the next day) is
// matched correctly — including against slots booked on the following date.
//
// DJ double-booking is checked first to match the API contract, so a candidate
// that clashes both by DJ and by stage reports "dj_double_booked".
func CheckConflicts(ctx context.Context, repo slotLister, candidate model.Slot, excludeSlotID, organizerID string) error {
	stored, err := repo.List(ctx, candidate.EventID, organizerID)
	if err != nil {
		return err
	}

	// Pass 1: DJ double-booked, any stage.
	if candidate.DjID != "" {
		for _, s := range stored {
			if s.ID == excludeSlotID {
				continue
			}
			if s.DjID == candidate.DjID && overlaps(candidate, s) {
				return &apperrors.ConflictError{Type: "dj_double_booked", ConflictingSlotID: s.ID}
			}
		}
	}

	// Pass 2: overlap on the same stage.
	for _, s := range stored {
		if s.ID == excludeSlotID {
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
// A slot whose times or date can't be parsed never overlaps (the database layer
// rejects the malformed write).
func overlaps(a, b model.Slot) bool {
	aStart, aEnd, aOK := absoluteInterval(a)
	bStart, bEnd, bOK := absoluteInterval(b)
	if !aOK || !bOK {
		return false
	}
	return aStart < bEnd && bStart < aEnd
}

// absoluteInterval returns a slot's [start, end) in minutes from a fixed epoch,
// treating an end at or before the start as running into the next day.
func absoluteInterval(s model.Slot) (start, end int, ok bool) {
	startTOD, endTOD := minutes(s.StartTime), minutes(s.EndTime)
	day, dayOK := dayOrdinal(s.SlotDate)
	if startTOD < 0 || endTOD < 0 || !dayOK {
		return 0, 0, false
	}
	start = day*minutesPerDay + startTOD
	dur := endTOD - startTOD
	if dur <= 0 {
		dur += minutesPerDay
	}
	return start, start + dur, true
}

// dayOrdinal converts a "YYYY-MM-DD" date to a day count from the Unix epoch.
func dayOrdinal(date string) (int, bool) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return 0, false
	}
	return int(t.Unix() / 86400), true
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
