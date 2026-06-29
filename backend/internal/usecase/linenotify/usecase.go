// Package linenotify holds the use case for managing a per-event LINE Notify
// token: encrypting it at rest and toggling whether notifications are enabled
// (US-006).
package linenotify

import (
	"context"

	"eventlineup/internal/domain/repository"
	"eventlineup/internal/infrastructure/crypto"
)

// UseCase encrypts LINE Notify tokens and persists them via the event repo.
type UseCase struct {
	repo repository.EventRepository
	key  []byte
}

// New builds the use case with the AES-256 key used to encrypt tokens at rest.
func New(repo repository.EventRepository, key []byte) *UseCase {
	return &UseCase{repo: repo, key: key}
}

// SetToken stores an encrypted LINE Notify token for the organizer's event, or
// clears it when rawToken is empty (disabling notifications). It returns whether
// LINE Notify is now enabled. A non-owned event surfaces as apperrors.ErrNotFound
// from the repository.
func (uc *UseCase) SetToken(ctx context.Context, eventID, organizerID, rawToken string) (bool, error) {
	if rawToken == "" {
		return uc.repo.SetLineNotifyToken(ctx, eventID, organizerID, nil)
	}
	enc, err := crypto.Encrypt(rawToken, uc.key)
	if err != nil {
		return false, err
	}
	return uc.repo.SetLineNotifyToken(ctx, eventID, organizerID, &enc)
}
