package dj_test

import (
	"context"
	"testing"
	"time"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
	dj "eventlineup/internal/usecase/dj"
)

// fakeDJRepo is a no-op DJRepository that records the expiry handed to
// SetPortalToken, so the token TTL can be asserted without a database.
type fakeDJRepo struct {
	gotHash      string
	gotExpiresAt time.Time
}

func (f *fakeDJRepo) List(context.Context, string, repository.DJListFilter) ([]model.DJ, error) {
	return nil, nil
}
func (f *fakeDJRepo) Get(context.Context, string, string) (model.DJ, error) { return model.DJ{}, nil }
func (f *fakeDJRepo) Create(context.Context, string, []string, string) (model.DJ, error) {
	return model.DJ{}, nil
}
func (f *fakeDJRepo) Update(context.Context, model.DJ, string) (model.DJ, error) {
	return model.DJ{}, nil
}
func (f *fakeDJRepo) Delete(context.Context, string, string) error { return nil }
func (f *fakeDJRepo) SetPortalToken(_ context.Context, _, _, tokenHash string, expiresAt time.Time) error {
	f.gotHash = tokenHash
	f.gotExpiresAt = expiresAt
	return nil
}
func (f *fakeDJRepo) GetByPortalToken(context.Context, string) (model.DJ, error) {
	return model.DJ{}, nil
}
func (f *fakeDJRepo) ListSlotsForDJ(context.Context, string) ([]model.PortalSlot, error) {
	return nil, nil
}

// Student Google login methods (migration 014) — unused by these tests.
func (f *fakeDJRepo) FindByGoogleID(context.Context, string) (model.DJ, error) {
	return model.DJ{}, nil
}
func (f *fakeDJRepo) FindByEmail(context.Context, string) (model.DJ, error) {
	return model.DJ{}, nil
}
func (f *fakeDJRepo) SetGoogleID(context.Context, string, string) error { return nil }
func (f *fakeDJRepo) SetEmail(context.Context, string, string, string) error {
	return nil
}

// TestGeneratePortalTokenTTL asserts a freshly minted token expires ~2 weeks out
// (EL-088 — shortened from 90 days), and that the returned expiry matches what
// was persisted.
func TestGeneratePortalTokenTTL(t *testing.T) {
	repo := &fakeDJRepo{}
	uc := dj.New(repo)

	before := time.Now()
	raw, expiresAt, err := uc.GeneratePortalToken(context.Background(), "dj-1", "org-1")
	if err != nil {
		t.Fatalf("GeneratePortalToken: %v", err)
	}
	if raw == "" {
		t.Fatal("expected a non-empty raw token")
	}

	const wantTTL = 14 * 24 * time.Hour
	gotTTL := expiresAt.Sub(before)
	if gotTTL < wantTTL-time.Minute || gotTTL > wantTTL+time.Minute {
		t.Errorf("token TTL = %v, want ~%v (2 weeks)", gotTTL, wantTTL)
	}
	if !repo.gotExpiresAt.Equal(expiresAt) {
		t.Errorf("expiry persisted (%v) != returned (%v)", repo.gotExpiresAt, expiresAt)
	}
}
