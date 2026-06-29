package dj

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"eventlineup/internal/domain/model"
)

// portalTokenTTL is how long a freshly generated DJ portal token stays valid.
const portalTokenTTL = 90 * 24 * time.Hour

// GeneratePortalToken issues a new portal token for a DJ the organizer owns. A
// 32-byte random token is generated; only its SHA-256 hash is stored (with a
// 90-day expiry). The raw token is returned exactly once — it is never persisted
// or logged. Generating a new token overwrites the old hash, which revokes any
// previously issued token. Returns apperrors.ErrNotFound if the DJ is not the
// organizer's.
func (uc *UseCase) GeneratePortalToken(ctx context.Context, djID, organizerID string) (rawToken string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	rawToken = hex.EncodeToString(buf)
	expiresAt = time.Now().Add(portalTokenTTL)
	if err = uc.repo.SetPortalToken(ctx, djID, organizerID, hashToken(rawToken), expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return rawToken, expiresAt, nil
}

// PortalByToken resolves a raw portal token to its DJ and that DJ's slots across
// all events. Returns apperrors.ErrNotFound if the token is unknown or expired.
func (uc *UseCase) PortalByToken(ctx context.Context, rawToken string) (model.DJ, []model.PortalSlot, error) {
	d, err := uc.repo.GetByPortalToken(ctx, hashToken(rawToken))
	if err != nil {
		return model.DJ{}, nil, err
	}
	slots, err := uc.repo.ListSlotsForDJ(ctx, d.ID)
	if err != nil {
		return model.DJ{}, nil, err
	}
	return d, slots, nil
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
