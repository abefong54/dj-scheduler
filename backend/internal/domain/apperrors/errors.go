package apperrors

import "errors"

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is the sentinel for scheduling conflicts. It lets callers use
// errors.Is(err, apperrors.ErrConflict) regardless of the concrete details.
var ErrConflict = errors.New("conflict")

// ConflictError describes a scheduling conflict that prevents a slot from being
// saved. Type is "dj_double_booked" or "stage_overlap" and ConflictingSlotID is
// the ID of the already-stored slot that the candidate collides with.
type ConflictError struct {
	Type              string
	ConflictingSlotID string
}

func (e *ConflictError) Error() string {
	return "conflict: " + e.Type + " (conflicting slot " + e.ConflictingSlotID + ")"
}

// Is reports ErrConflict so errors.Is(err, apperrors.ErrConflict) matches any
// ConflictError while errors.As recovers the structured details.
func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}
