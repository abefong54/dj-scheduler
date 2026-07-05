// Package auth handles organizer authentication via external OAuth providers:
// it turns an OAuth callback into an upserted organizer and a signed JWT. The
// provider (Google today, LINE next) is abstracted behind OAuthProvider so both
// flow through one callback → upsert → token path.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
	"eventlineup/internal/token"
)

// Provider identifiers. Each maps to a registered OAuthProvider and to the
// column an organizer's external identity is stored under.
const (
	ProviderGoogle = "google"
)

// ExternalUser is the minimal profile an OAuth provider returns. Subject is the
// provider's stable per-user id (Google's `id`, LINE's `sub`).
type ExternalUser struct {
	Provider string
	Subject  string
	Email    string
	Name     string
}

// OAuthProvider abstracts one provider's OAuth dance so the usecase stays free
// of net/http and the oauth2 library. Real implementations live under
// internal/infrastructure (e.g. googleauth).
type OAuthProvider interface {
	AuthCodeURL(state string) string
	UserFromCode(ctx context.Context, code string) (ExternalUser, error)
}

type UseCase struct {
	providers   map[string]OAuthProvider
	repo        repository.OrganizerRepository
	memberships repository.MembershipRepository
	jwtSecret   string
	tokenTTL    time.Duration
}

func New(providers map[string]OAuthProvider, repo repository.OrganizerRepository, memberships repository.MembershipRepository, jwtSecret string, tokenTTL time.Duration) *UseCase {
	return &UseCase{providers: providers, repo: repo, memberships: memberships, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

// AuthCodeURL returns the consent-screen URL for the named provider to redirect
// the user to, or an error if the provider isn't registered.
func (uc *UseCase) AuthCodeURL(provider, state string) (string, error) {
	p, ok := uc.providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown auth provider %q", provider)
	}
	return p.AuthCodeURL(state), nil
}

// HandleCallback exchanges an OAuth code for the provider's profile, upserts the
// matching organizer, and returns a signed JWT.
func (uc *UseCase) HandleCallback(ctx context.Context, provider, code string) (string, error) {
	p, ok := uc.providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown auth provider %q", provider)
	}
	eu, err := p.UserFromCode(ctx, code)
	if err != nil {
		return "", err
	}
	org, err := uc.upsertOrganizer(ctx, eu)
	if err != nil {
		return "", err
	}
	// Invite auto-link (EL-104): if this person was invited to a school by email,
	// attach the pending membership to their account on sign-in. Best-effort — a
	// failure here shouldn't block login, so the error is intentionally ignored.
	_ = uc.memberships.LinkPendingByEmail(ctx, org.ID, org.Email)
	return token.Sign(uc.jwtSecret, org.ID, org.Email, uc.tokenTTL)
}

// upsertOrganizer returns the existing organizer for this external identity,
// creating one on first sign-in. Re-signing in never creates a duplicate.
func (uc *UseCase) upsertOrganizer(ctx context.Context, eu ExternalUser) (model.Organizer, error) {
	switch eu.Provider {
	case ProviderGoogle:
		org, err := uc.repo.FindByGoogleID(ctx, eu.Subject)
		if errors.Is(err, apperrors.ErrNotFound) {
			return uc.repo.Create(ctx, eu.Email, eu.Name, eu.Subject)
		}
		if err != nil {
			return model.Organizer{}, err
		}
		return org, nil
	default:
		return model.Organizer{}, fmt.Errorf("unsupported auth provider %q", eu.Provider)
	}
}
