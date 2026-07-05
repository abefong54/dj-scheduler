package submission

import (
	"context"
	"errors"
	"testing"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

// fakeRepo is an in-memory SubmissionRepository for usecase tests (no DB).
type fakeRepo struct {
	upserted    model.Submission
	reviewedArg struct{ id, status, note, badgeLabel, badgeIcon string }
}

func (f *fakeRepo) Create(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error) {
	return s, nil
}
func (f *fakeRepo) Get(ctx context.Context, id, schoolID string) (model.Submission, error) {
	return model.Submission{}, apperrors.ErrNotFound
}
func (f *fakeRepo) UpsertForChallenge(ctx context.Context, s model.Submission, schoolID string) (model.Submission, error) {
	s.SchoolID = schoolID
	s.Status = "posted"
	f.upserted = s
	return s, nil
}
func (f *fakeRepo) Review(ctx context.Context, id, schoolID, status, note, badgeLabel, badgeIcon string) (model.Submission, error) {
	f.reviewedArg = struct{ id, status, note, badgeLabel, badgeIcon string }{id, status, note, badgeLabel, badgeIcon}
	return model.Submission{ID: id, Status: status}, nil
}
func (f *fakeRepo) ListFeed(ctx context.Context, schoolID string) ([]model.Submission, error) {
	return nil, nil
}
func (f *fakeRepo) ListForStudent(ctx context.Context, djID, schoolID string) ([]model.Submission, error) {
	return nil, nil
}

func TestPostResolvesMediaAndUpserts(t *testing.T) {
	f := &fakeRepo{}
	uc := New(f)
	got, err := uc.Post(context.Background(), "dj-1", "school-1", "chal-1", "https://youtu.be/abc123", "my first mix")
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if got.MediaKind != "embed" || got.MediaProvider != "youtube" || got.MediaRef != "abc123" {
		t.Fatalf("media not resolved onto submission: %+v", got)
	}
	if f.upserted.DJID != "dj-1" || f.upserted.ChallengeID != "chal-1" || f.upserted.Caption != "my first mix" {
		t.Fatalf("upsert got wrong submission: %+v", f.upserted)
	}
	if got.Status != "posted" {
		t.Fatalf("want status posted, got %q", got.Status)
	}
}

func TestPostRejectsBadLink(t *testing.T) {
	uc := New(&fakeRepo{})
	_, err := uc.Post(context.Background(), "dj-1", "school-1", "chal-1", "https://vimeo.com/1", "")
	var v *apperrors.ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestPostRequiresChallenge(t *testing.T) {
	uc := New(&fakeRepo{})
	_, err := uc.Post(context.Background(), "dj-1", "school-1", "", "https://youtu.be/abc", "")
	var v *apperrors.ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("want validation error for missing challenge, got %v", err)
	}
}

func TestReviewValidatesStatus(t *testing.T) {
	f := &fakeRepo{}
	uc := New(f)

	if _, err := uc.Review(context.Background(), "sub-1", "school-1", "approved", "nice"); err == nil {
		t.Fatal("want validation error for bad status")
	}
	if _, err := uc.Review(context.Background(), "sub-1", "school-1", "passed", "great work"); err != nil {
		t.Fatalf("passed should be valid: %v", err)
	}
	if f.reviewedArg.status != "passed" || f.reviewedArg.id != "sub-1" {
		t.Fatalf("review not forwarded: %+v", f.reviewedArg)
	}
}
