// Package studentauth handles student authentication via Google SSO. It parallels
// usecase/auth (organizer) but resolves a student (a model.DJ) by google_id, then
// by email for the first-login auto-link, and rejects Google users not on any
// roster. SCAFFOLD: shapes only — Track A (login backend) fills the resolution
// path. See 2026-07-05-student-google-login-design.
package studentauth

import (
	"context"
	"errors"
	"time"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
	"eventlineup/internal/token"
)

// GoogleUser is the minimal Google profile the usecase needs.
type GoogleUser struct {
	GoogleID string
	Email    string
	Name     string
}

// GoogleAuthenticator abstracts the Google OAuth dance (same shape as
// usecase/auth). The real implementation lives in infrastructure/googleauth.
type GoogleAuthenticator interface {
	AuthCodeURL(state string) string
	UserFromCode(ctx context.Context, code string) (GoogleUser, error)
}

type UseCase struct {
	google    GoogleAuthenticator
	djs       repository.DJRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func New(google GoogleAuthenticator, djs repository.DJRepository, jwtSecret string, tokenTTL time.Duration) *UseCase {
	return &UseCase{google: google, djs: djs, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

// AuthCodeURL returns the Google consent-screen URL to redirect the student to.
func (uc *UseCase) AuthCodeURL(state string) string {
	return uc.google.AuthCodeURL(state)
}

// HandleCallback exchanges an OAuth code for a student session JWT. It resolves
// the student by google_id (returning), then by email (first login → auto-link),
// and returns apperrors.ErrNotFound when the Google user is on no roster (handler
// maps to 403). The returned JWT is scoped to the student's DJ id and school.
func (uc *UseCase) HandleCallback(ctx context.Context, code string) (string, error) {
	gu, err := uc.google.UserFromCode(ctx, code)
	if err != nil {
		return "", err
	}
	dj, err := uc.resolveStudent(ctx, gu)
	if err != nil {
		return "", err
	}
	return token.SignStudent(uc.jwtSecret, dj.ID, dj.SchoolID, uc.tokenTTL)
}

// resolveStudent maps a Google profile to a roster DJ. A returning student is
// found by google_id; a first-time student is matched by their teacher-set email
// and has their google_id bound (auto-link). A Google user on no roster yields
// apperrors.ErrNotFound — enrollment is teacher-controlled, so there is no
// auto-create.
func (uc *UseCase) resolveStudent(ctx context.Context, gu GoogleUser) (model.DJ, error) {
	dj, err := uc.djs.FindByGoogleID(ctx, gu.GoogleID)
	if err == nil {
		return dj, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return model.DJ{}, err
	}

	// First login: match the teacher-provisioned roster email, then bind the
	// Google identity so subsequent logins resolve by google_id.
	dj, err = uc.djs.FindByEmail(ctx, gu.Email)
	if err != nil {
		return model.DJ{}, err // ErrNotFound bubbles up → handler maps to 403.
	}
	if err := uc.djs.SetGoogleID(ctx, dj.ID, gu.GoogleID); err != nil {
		return model.DJ{}, err
	}
	dj.GoogleID = gu.GoogleID
	return dj, nil
}
