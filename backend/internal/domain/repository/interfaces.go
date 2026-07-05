package repository

import (
	"context"
	"time"

	"eventlineup/internal/domain/model"
)

// DJListFilter narrows the DJ list (EL-019). The zero value lists everyone.
type DJListFilter struct {
	// CertifiedFor, when non-empty, keeps only DJs certified for that genre
	// (case-insensitive). ReadyOnly keeps only DJs with at least one certification.
	CertifiedFor string
	ReadyOnly    bool
}

// DJRepository is scoped to the caller's school (EL-104): DJs are school-owned.
type DJRepository interface {
	List(ctx context.Context, schoolID string, filter DJListFilter) ([]model.DJ, error)
	Get(ctx context.Context, id, schoolID string) (model.DJ, error)
	Create(ctx context.Context, name string, tags []string, schoolID string) (model.DJ, error)
	Update(ctx context.Context, dj model.DJ, schoolID string) (model.DJ, error)
	Delete(ctx context.Context, id, schoolID string) error

	// DJ portal tokens (US-009).
	SetPortalToken(ctx context.Context, djID, schoolID, tokenHash string, expiresAt time.Time) error
	GetByPortalToken(ctx context.Context, tokenHash string) (model.DJ, error)
	ListSlotsForDJ(ctx context.Context, djID string) ([]model.PortalSlot, error)

	// Student Google login (migration 014). FindByGoogleID/FindByEmail are the
	// two-step resolution path (returning student, then first-login link) and are
	// intentionally unscoped: identity is resolved before a school is known — the
	// student's school comes from the matched DJ row. Both return
	// apperrors.ErrNotFound when no student matches. SetGoogleID binds the Google
	// identity on first login; SetEmail is the teacher provisioning path (scoped).
	FindByGoogleID(ctx context.Context, googleID string) (model.DJ, error)
	FindByEmail(ctx context.Context, email string) (model.DJ, error)
	SetGoogleID(ctx context.Context, djID, googleID string) error
	SetEmail(ctx context.Context, djID, schoolID, email string) error
}

// ChallengeRepository is scoped to the caller's school (challenges are
// school-owned), mirroring SlotRepository/PerformanceRepository. A challenge that
// isn't the school's reads as apperrors.ErrNotFound.
type ChallengeRepository interface {
	// List returns every non-archived challenge (core + custom) for the school.
	List(ctx context.Context, schoolID string) ([]model.Challenge, error)
	Get(ctx context.Context, id, schoolID string) (model.Challenge, error)
	Create(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error)
	Update(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error)
	// Archive soft-deletes a challenge (archived=true); earned badges survive.
	Archive(ctx context.Context, id, schoolID string) error
	// ListCore returns the school's core (badge-bearing) challenges for "Browse Core".
	ListCore(ctx context.Context, schoolID string) ([]model.Challenge, error)
	// ListForStudent returns the challenges relevant to one student: the core set
	// plus any custom challenges assigned to them.
	ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Challenge, error)
}

// SubmissionRepository is scoped to the caller's school (submissions are
// school-owned), mirroring SlotRepository. A submission that isn't the school's
// reads as apperrors.ErrNotFound.
type SubmissionRepository interface {
	Create(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error)
	Get(ctx context.Context, id, schoolID string) (model.Submission, error)
	// UpsertForChallenge writes the one active submission per (dj_id, challenge_id):
	// re-work updates the existing row rather than appending history.
	UpsertForChallenge(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error)
	// Review records a teacher pass/needs_work with an optional note. On a pass of
	// a core challenge the badge is snapshotted onto the submission (badgeLabel,
	// badgeIcon); pass empty strings for needs_work or non-core passes.
	Review(ctx context.Context, id, schoolID, status, note, badgeLabel, badgeIcon string) (model.Submission, error)
	// ListFeed returns the school's passed submissions, newest first (the feed).
	ListFeed(ctx context.Context, schoolID string) ([]model.Submission, error)
	// ListForStudent returns one student's submissions across all their challenges.
	ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Submission, error)
}

// EventRepository is scoped to the caller's school (EL-104). Create/Clone also
// record createdBy (the organizer who scheduled it) as an audit column.
type EventRepository interface {
	List(ctx context.Context, schoolID string) ([]model.Event, error)
	Get(ctx context.Context, id, schoolID string) (model.Event, error)
	// GetPublic reads an event without school scoping, for the public schedule.
	GetPublic(ctx context.Context, id string) (model.Event, error)
	Create(ctx context.Context, e model.Event, createdBy, schoolID string) (model.Event, error)
	Update(ctx context.Context, e model.Event, schoolID string) (model.Event, error)
	Delete(ctx context.Context, id, schoolID string) error
	Clone(ctx context.Context, id, createdBy, schoolID string) (model.Event, error)
	// SetLineNotifyToken stores encToken (an already-encrypted token) for the
	// school's event, or clears it when encToken is nil. It returns whether LINE
	// Notify is now enabled, and apperrors.ErrNotFound if the event is not the
	// school's. See US-006.
	SetLineNotifyToken(ctx context.Context, id, schoolID string, encToken *string) (bool, error)
}

// StageRepository scopes every operation to the school that owns the parent
// event (EL-104); a stage that isn't theirs reads as not-found.
type StageRepository interface {
	List(ctx context.Context, eventID, schoolID string) ([]model.Stage, error)
	// ListPublic is unscoped, for the public schedule endpoint only.
	ListPublic(ctx context.Context, eventID string) ([]model.Stage, error)
	Get(ctx context.Context, id, eventID, schoolID string) (model.Stage, error)
	Create(ctx context.Context, eventID, name, color, schoolID string) (model.Stage, error)
	Update(ctx context.Context, s model.Stage, schoolID string) (model.Stage, error)
	Delete(ctx context.Context, id, eventID, schoolID string) error
}

// PerformanceFilter narrows a roster-wide performance summary to a window and/or
// a single event (EL-043). Empty fields mean "no bound".
type PerformanceFilter struct {
	EventID string // restrict to one event; "" = all events
	From    string // inclusive slot_date lower bound "YYYY-MM-DD"; "" = no lower bound
	To      string // inclusive slot_date upper bound; "" = no upper bound
}

// PerformanceRepository aggregates played slots into performance reps, always
// scoped to the school that owns the parent events (EL-104/EL-043).
type PerformanceRepository interface {
	// DJPerformance aggregates one DJ's reps across the school's events with a
	// by-genre breakdown. Returns apperrors.ErrNotFound if the DJ isn't the
	// school's; an owned DJ who never played returns a zero-rep summary.
	DJPerformance(ctx context.Context, djID, schoolID string) (model.DJPerformance, error)
	// RosterSummary returns every active student (is_student=true) of the school
	// with their rep count in the window, including zero-rep students.
	RosterSummary(ctx context.Context, schoolID string, filter PerformanceFilter) ([]model.RosterPerformance, error)
}

type LeadRepository interface {
	Create(ctx context.Context, lead model.Lead) (model.Lead, error)
}

type OrganizerRepository interface {
	FindByGoogleID(ctx context.Context, googleID string) (model.Organizer, error)
	FindByLineID(ctx context.Context, lineID string) (model.Organizer, error)
	FindByEmail(ctx context.Context, email string) (model.Organizer, error)
	FindByID(ctx context.Context, id string) (model.Organizer, error)
	Create(ctx context.Context, email, name, googleID string) (model.Organizer, error)
	CreateWithLineID(ctx context.Context, email, name, lineID string) (model.Organizer, error)
	LinkLineID(ctx context.Context, organizerID, lineID string) error
}

// SchoolRepository manages tenant schools — the owner of events and DJs (EL-103).
type SchoolRepository interface {
	Create(ctx context.Context, name, city, createdBy string) (model.School, error)
	Get(ctx context.Context, id string) (model.School, error)
}

// MembershipRepository links organizers to schools (EL-103). A pending
// membership has a NULL organizer_id until the invited person signs in.
type MembershipRepository interface {
	Create(ctx context.Context, schoolID string, organizerID *string, email, role, status string) (model.Membership, error)
	// ActiveForOrganizer returns the organizer's active membership (their current
	// school), or apperrors.ErrNotFound if they have none yet.
	ActiveForOrganizer(ctx context.Context, organizerID string) (model.Membership, error)
	ListBySchool(ctx context.Context, schoolID string) ([]model.Membership, error)
	// LinkPendingByEmail attaches a pending membership to organizerID on their
	// first matching sign-in (email match) and activates it.
	LinkPendingByEmail(ctx context.Context, organizerID, email string) error
}

// SlotRepository scopes every operation to the school that owns the parent
// event (EL-104), except SetDJConfirmation which is scoped to the DJ instead.
type SlotRepository interface {
	List(ctx context.Context, eventID, schoolID string) ([]model.Slot, error)
	// ListPublic is unscoped, for the public schedule endpoint only.
	ListPublic(ctx context.Context, eventID string) ([]model.Slot, error)
	// GetPublicByID looks up a slot by id alone, without school/event scoping,
	// for the public per-DJ share card (EL-049).
	GetPublicByID(ctx context.Context, id string) (model.Slot, error)
	Get(ctx context.Context, id, eventID, schoolID string) (model.Slot, error)
	Create(ctx context.Context, s model.Slot, eventID, schoolID string) (model.Slot, error)
	Update(ctx context.Context, s model.Slot, eventID, schoolID string) (model.Slot, error)
	Delete(ctx context.Context, id, eventID, schoolID string) error
	// SetDJConfirmation records a DJ portal confirm/flag, scoped to the DJ who owns
	// the slot (ErrForbidden when the slot isn't theirs). US-011.
	SetDJConfirmation(ctx context.Context, slotID, djID, confirmation string) error
}
