// Package auth handles organizer authentication via Google SSO: it turns an
// OAuth callback into an upserted organizer and a signed JWT.
package auth

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

// GoogleAuthenticator abstracts the Google OAuth dance so the usecase stays
// free of net/http and the oauth2 library. The real implementation lives in
// internal/infrastructure/googleauth.
type GoogleAuthenticator interface {
	AuthCodeURL(state string) string
	UserFromCode(ctx context.Context, code string) (GoogleUser, error)
}

type UseCase struct {
	google    GoogleAuthenticator
	repo      repository.OrganizerRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func New(google GoogleAuthenticator, repo repository.OrganizerRepository, jwtSecret string, tokenTTL time.Duration) *UseCase {
	return &UseCase{google: google, repo: repo, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

// AuthCodeURL returns the Google consent-screen URL to redirect the user to.
func (uc *UseCase) AuthCodeURL(state string) string {
	return uc.google.AuthCodeURL(state)
}

// HandleCallback exchanges an OAuth code for a Google profile, upserts the
// matching organizer, and returns a signed JWT.
func (uc *UseCase) HandleCallback(ctx context.Context, code string) (string, error) {
	gu, err := uc.google.UserFromCode(ctx, code)
	if err != nil {
		return "", err
	}
	org, err := uc.upsertOrganizer(ctx, gu)
	if err != nil {
		return "", err
	}
	return token.Sign(uc.jwtSecret, org.ID, org.Email, uc.tokenTTL)
}

// upsertOrganizer returns the existing organizer for this Google ID, creating
// one on first sign-in. Re-signing in never creates a duplicate.
func (uc *UseCase) upsertOrganizer(ctx context.Context, gu GoogleUser) (model.Organizer, error) {
	org, err := uc.repo.FindByGoogleID(ctx, gu.GoogleID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return uc.repo.Create(ctx, gu.Email, gu.Name, gu.GoogleID)
	}
	if err != nil {
		return model.Organizer{}, err
	}
	return org, nil
}
