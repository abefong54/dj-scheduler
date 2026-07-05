// Package school handles the tenant onboarding surface (EL-105): who am I, which
// school do I belong to, register a school, and invite employees.
package school

import (
	"context"
	"errors"
	"strings"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	organizers  repository.OrganizerRepository
	schools     repository.SchoolRepository
	memberships repository.MembershipRepository
}

func New(o repository.OrganizerRepository, s repository.SchoolRepository, m repository.MembershipRepository) *UseCase {
	return &UseCase{organizers: o, schools: s, memberships: m}
}

// SchoolMembership pairs the caller's school with their role in it.
type SchoolMembership struct {
	School model.School
	Role   string
}

// Me returns the caller's user record and their active school (nil if they
// haven't joined/registered one yet — the onboarding entry state).
func (uc *UseCase) Me(ctx context.Context, organizerID string) (model.Organizer, *SchoolMembership, error) {
	org, err := uc.organizers.FindByID(ctx, organizerID)
	if err != nil {
		return model.Organizer{}, nil, err
	}
	m, err := uc.memberships.ActiveForOrganizer(ctx, organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return org, nil, nil
	}
	if err != nil {
		return model.Organizer{}, nil, err
	}
	school, err := uc.schools.Get(ctx, m.SchoolID)
	if err != nil {
		return model.Organizer{}, nil, err
	}
	return org, &SchoolMembership{School: school, Role: m.Role}, nil
}

// Register creates a school owned by the caller with an owner membership. Returns
// a validation error if the caller already belongs to a school (one per user for
// the demo).
func (uc *UseCase) Register(ctx context.Context, organizerID, name, city string) (model.School, string, error) {
	if _, err := uc.memberships.ActiveForOrganizer(ctx, organizerID); err == nil {
		return model.School{}, "", apperrors.NewValidation("you already belong to a school")
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return model.School{}, "", err
	}
	org, err := uc.organizers.FindByID(ctx, organizerID)
	if err != nil {
		return model.School{}, "", err
	}
	school, err := uc.schools.Create(ctx, name, city, organizerID)
	if err != nil {
		return model.School{}, "", err
	}
	if _, err := uc.memberships.Create(ctx, school.ID, &organizerID, org.Email, "owner", "active"); err != nil {
		return model.School{}, "", err
	}
	return school, "owner", nil
}

func (uc *UseCase) ListMembers(ctx context.Context, schoolID string) ([]model.Membership, error) {
	return uc.memberships.ListBySchool(ctx, schoolID)
}

// Invite adds a pending membership by email; on the invitee's first matching
// sign-in it becomes active (EL-104 auto-link). A duplicate email in the school
// is rejected.
func (uc *UseCase) Invite(ctx context.Context, schoolID, email, role string) (model.Membership, error) {
	if role != "owner" && role != "admin" && role != "member" {
		return model.Membership{}, apperrors.NewValidation("role must be owner, admin, or member")
	}
	existing, err := uc.memberships.ListBySchool(ctx, schoolID)
	if err != nil {
		return model.Membership{}, err
	}
	for _, m := range existing {
		if strings.EqualFold(m.Email, email) {
			return model.Membership{}, apperrors.NewValidation("that email is already a member")
		}
	}
	return uc.memberships.Create(ctx, schoolID, nil, email, role, "pending")
}
