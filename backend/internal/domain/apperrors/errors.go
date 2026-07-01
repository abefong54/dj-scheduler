package apperrors

import "errors"

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrForbidden is returned when the caller is authenticated but not permitted to
// act on the resource (e.g. a DJ portal token targeting a slot that isn't theirs).
var ErrForbidden = errors.New("forbidden")

// ErrConflict is the sentinel for scheduling conflicts. It lets callers use
// errors.Is(err, apperrors.ErrConflict) regardless of the concrete details.
var ErrConflict = errors.New("conflict")

// ValidationError is a client-safe validation failure. Its Message is authored
// in code (a required-field or format complaint) — never derived from a driver
// or DB error — so it is safe to return to the client verbatim with a 400.
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

// NewValidation builds a ValidationError with the given client-safe message.
func NewValidation(msg string) *ValidationError { return &ValidationError{Message: msg} }

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
