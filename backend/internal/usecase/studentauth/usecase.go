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

	"eventlineup/internal/domain/repository"
)

var errNotImplemented = errors.New("not implemented")

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
// maps to 403). SCAFFOLD: not implemented.
func (uc *UseCase) HandleCallback(ctx context.Context, code string) (string, error) {
	return "", errNotImplemented
}
