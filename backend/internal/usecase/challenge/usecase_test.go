package challenge

import (
	"context"
	"errors"
	"testing"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

// fakeRepo is an in-memory ChallengeRepository for usecase tests (no DB).
type fakeRepo struct {
	created model.Challenge
	core    []model.Challenge
	custom  []model.Challenge
}

func (f *fakeRepo) List(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return nil, nil
}
func (f *fakeRepo) Get(ctx context.Context, id, schoolID string) (model.Challenge, error) {
	return model.Challenge{}, apperrors.ErrNotFound
}
func (f *fakeRepo) Create(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	f.created = c
	c.ID = "new"
	return c, nil
}
func (f *fakeRepo) Update(ctx context.Context, c model.Challenge, schoolID string) (model.Challenge, error) {
	return c, nil
}
func (f *fakeRepo) Archive(ctx context.Context, id, schoolID string) error { return nil }
func (f *fakeRepo) ListCore(ctx context.Context, schoolID string) ([]model.Challenge, error) {
	return f.core, nil
}
func (f *fakeRepo) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Challenge, error) {
	return f.custom, nil
}

func TestCreateEnforcesTypeInvariants(t *testing.T) {
	cases := []struct {
		name string
		c    model.Challenge
		ok   bool
	}{
		{"core ok", model.Challenge{Type: "core", Title: "T", BadgeLabel: "B"}, true},
		{"core missing badge", model.Challenge{Type: "core", Title: "T"}, false},
		{"core with assignee", model.Challenge{Type: "core", Title: "T", BadgeLabel: "B", AssignedDJID: "dj"}, false},
		{"custom ok", model.Challenge{Type: "custom", Title: "T", AssignedDJID: "dj"}, true},
		{"custom missing assignee", model.Challenge{Type: "custom", Title: "T"}, false},
		{"custom carrying badge", model.Challenge{Type: "custom", Title: "T", AssignedDJID: "dj", BadgeLabel: "B"}, false},
		{"missing title", model.Challenge{Type: "core", BadgeLabel: "B"}, false},
		{"bad type", model.Challenge{Type: "weird", Title: "T"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(&fakeRepo{}).Create(context.Background(), tc.c, "s1")
			if tc.ok {
				if err != nil {
					t.Fatalf("want ok, got %v", err)
				}
				return
			}
			var v *apperrors.ValidationError
			if !errors.As(err, &v) {
				t.Fatalf("want validation error, got %v", err)
			}
		})
	}
}

func TestUpdateAlsoValidates(t *testing.T) {
	_, err := New(&fakeRepo{}).Update(context.Background(), model.Challenge{Type: "core", Title: "T"}, "s1")
	var v *apperrors.ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("Update should re-validate; want validation error, got %v", err)
	}
}

func TestListForStudentMergesCoreThenCustom(t *testing.T) {
	f := &fakeRepo{
		core:   []model.Challenge{{ID: "c1", Type: "core"}, {ID: "c2", Type: "core"}},
		custom: []model.Challenge{{ID: "x1", Type: "custom"}},
	}
	got, err := New(f).ListForStudent(context.Background(), "dj1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].ID != "c1" || got[2].ID != "x1" {
		t.Fatalf("want core then custom [c1 c2 x1], got %+v", got)
	}
}
