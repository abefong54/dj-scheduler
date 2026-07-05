package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

// --- fakes (no DB) --------------------------------------------------------

type fakeProvider struct{ user ExternalUser }

func (f fakeProvider) AuthCodeURL(state string) string { return "https://provider/auth?state=" + state }
func (f fakeProvider) UserFromCode(context.Context, string) (ExternalUser, error) {
	return f.user, nil
}

type fakeOrgRepo struct {
	byGoogleID map[string]model.Organizer
	created    []model.Organizer
	nextID     int
}

func newFakeOrgRepo() *fakeOrgRepo {
	return &fakeOrgRepo{byGoogleID: map[string]model.Organizer{}}
}
func (r *fakeOrgRepo) FindByGoogleID(_ context.Context, gid string) (model.Organizer, error) {
	if o, ok := r.byGoogleID[gid]; ok {
		return o, nil
	}
	return model.Organizer{}, apperrors.ErrNotFound
}
func (r *fakeOrgRepo) FindByID(_ context.Context, id string) (model.Organizer, error) {
	for _, o := range r.byGoogleID {
		if o.ID == id {
			return o, nil
		}
	}
	return model.Organizer{}, apperrors.ErrNotFound
}
func (r *fakeOrgRepo) Create(_ context.Context, email, name, googleID string) (model.Organizer, error) {
	r.nextID++
	o := model.Organizer{ID: fmt.Sprintf("org-%d", r.nextID), Email: email, Name: name, GoogleID: googleID}
	r.byGoogleID[googleID] = o
	r.created = append(r.created, o)
	return o, nil
}
func (r *fakeOrgRepo) FindByLineID(_ context.Context, lineID string) (model.Organizer, error) {
	for _, o := range r.created {
		if lineID != "" && o.LineID == lineID {
			return o, nil
		}
	}
	return model.Organizer{}, apperrors.ErrNotFound
}
func (r *fakeOrgRepo) FindByEmail(_ context.Context, email string) (model.Organizer, error) {
	for _, o := range r.created {
		if o.Email == email {
			return o, nil
		}
	}
	return model.Organizer{}, apperrors.ErrNotFound
}
func (r *fakeOrgRepo) CreateWithLineID(_ context.Context, email, name, lineID string) (model.Organizer, error) {
	r.nextID++
	o := model.Organizer{ID: fmt.Sprintf("org-%d", r.nextID), Email: email, Name: name, LineID: lineID}
	r.created = append(r.created, o)
	return o, nil
}
func (r *fakeOrgRepo) LinkLineID(_ context.Context, organizerID, lineID string) error {
	for i := range r.created {
		if r.created[i].ID == organizerID {
			r.created[i].LineID = lineID
			return nil
		}
	}
	return apperrors.ErrNotFound
}

type fakeMemRepo struct{ linked int }

func (m *fakeMemRepo) Create(context.Context, string, *string, string, string, string) (model.Membership, error) {
	return model.Membership{}, nil
}
func (m *fakeMemRepo) ActiveForOrganizer(context.Context, string) (model.Membership, error) {
	return model.Membership{}, apperrors.ErrNotFound
}
func (m *fakeMemRepo) ListBySchool(context.Context, string) ([]model.Membership, error) {
	return nil, nil
}
func (m *fakeMemRepo) LinkPendingByEmail(context.Context, string, string) error {
	m.linked++
	return nil
}

// --- tests ----------------------------------------------------------------

func newUC(providers map[string]OAuthProvider, repo *fakeOrgRepo, mem *fakeMemRepo) *UseCase {
	return New(providers, repo, mem, "unit-secret-unit-secret-unit-secret-01", time.Hour)
}

// The shared, provider-parameterized path: a callback upserts the organizer for
// the provider's subject, attempts the invite-link, and signs a token — and a
// second callback with the same subject does not create a duplicate.
func TestHandleCallbackGoogleUpsertsAndSigns(t *testing.T) {
	repo, mem := newFakeOrgRepo(), &fakeMemRepo{}
	uc := newUC(map[string]OAuthProvider{
		ProviderGoogle: fakeProvider{user: ExternalUser{Provider: ProviderGoogle, Subject: "g-1", Email: "a@b.com", Name: "A"}},
	}, repo, mem)

	tok, err := uc.HandleCallback(context.Background(), ProviderGoogle, "code")
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if tok == "" {
		t.Fatal("expected a signed token")
	}
	if len(repo.created) != 1 || repo.created[0].GoogleID != "g-1" {
		t.Fatalf("expected one organizer created for g-1, got %+v", repo.created)
	}
	if mem.linked != 1 {
		t.Fatalf("expected one invite-link attempt, got %d", mem.linked)
	}

	if _, err := uc.HandleCallback(context.Background(), ProviderGoogle, "code"); err != nil {
		t.Fatalf("second callback: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("re-signin must not duplicate; created=%d", len(repo.created))
	}
}

func TestHandleCallbackUnknownProvider(t *testing.T) {
	uc := newUC(map[string]OAuthProvider{}, newFakeOrgRepo(), &fakeMemRepo{})
	if _, err := uc.HandleCallback(context.Background(), "line", "code"); err == nil {
		t.Fatal("expected an error for an unregistered provider")
	}
}

func TestAuthCodeURLUnknownProvider(t *testing.T) {
	uc := newUC(map[string]OAuthProvider{}, newFakeOrgRepo(), &fakeMemRepo{})
	if _, err := uc.AuthCodeURL("line", "state"); err == nil {
		t.Fatal("expected an error for an unregistered provider")
	}
}
