package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	submissionuc "eventlineup/internal/usecase/submission"
)

func submissionStudentRouter(pool *pgxpool.Pool, schoolID, djID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(studentContext(schoolID, djID))
	httphandler.NewSubmissionHandler(submissionuc.New(database.NewSubmissionRepository(pool))).
		RegisterStudent(r.Group("/api"))
	return r
}

func submissionTeacherRouter(pool *pgxpool.Pool, org string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(orgContext(pool, org))
	httphandler.NewSubmissionHandler(submissionuc.New(database.NewSubmissionRepository(pool))).
		RegisterProtected(r.Group("/api"))
	return r
}

func seedCoreChallenge(t *testing.T, pool *pgxpool.Pool, schoolID, org, badgeLabel string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO challenges (school_id, type, title, badge_label, badge_icon, sort_order, created_by)
		 VALUES ($1,'core','Test Core',$2,'🥁',10,$3) RETURNING id`,
		schoolID, badgeLabel, org).Scan(&id); err != nil {
		t.Fatalf("seedCoreChallenge: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM challenges WHERE id = $1", id) })
	return id
}

func TestSubmissionPostReviewFeedFlow(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	schoolID := schoolIDForOrg(t, pool, org)
	dj := seedDJ(t, pool, org, "poster")
	chal := seedCoreChallenge(t, pool, schoolID, org, "Beatmatcher")

	sr := submissionStudentRouter(pool, schoolID, dj)

	// Post an attempt with a valid YouTube link.
	w := jsonReq(t, sr, http.MethodPost, "/api/student/submissions", map[string]any{
		"challenge_id": chal, "media_url": "https://youtu.be/abc123XYZ", "caption": "my mix",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("post attempt: want 201, got %d (%s)", w.Code, w.Body.String())
	}
	var sub model.Submission
	json.Unmarshal(w.Body.Bytes(), &sub)
	if sub.ID == "" || sub.MediaProvider != "youtube" || sub.MediaRef != "abc123XYZ" || sub.Status != "posted" {
		t.Fatalf("unexpected submission: %+v", sub)
	}

	// Feed is empty before a pass.
	w = jsonReq(t, sr, http.MethodGet, "/api/student/feed", nil)
	var feed []model.Submission
	json.Unmarshal(w.Body.Bytes(), &feed)
	if len(feed) != 0 {
		t.Fatalf("feed should be empty before pass, got %+v", feed)
	}

	// Teacher passes it → badge snapshotted (core challenge).
	tr := submissionTeacherRouter(pool, org)
	w = jsonReq(t, tr, http.MethodPost, "/api/submissions/"+sub.ID+"/review", map[string]any{
		"status": "passed", "review_note": "great work",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("review pass: want 200, got %d (%s)", w.Code, w.Body.String())
	}
	var reviewed model.Submission
	json.Unmarshal(w.Body.Bytes(), &reviewed)
	if reviewed.Status != "passed" {
		t.Fatalf("want status passed, got %q", reviewed.Status)
	}
	if reviewed.EarnedBadgeLabel != "Beatmatcher" {
		t.Fatalf("core pass should snapshot the badge, got %q", reviewed.EarnedBadgeLabel)
	}

	// Feed now includes the passed submission.
	w = jsonReq(t, sr, http.MethodGet, "/api/student/feed", nil)
	json.Unmarshal(w.Body.Bytes(), &feed)
	if len(feed) != 1 || feed[0].ID != sub.ID {
		t.Fatalf("passed submission should be in the feed, got %+v", feed)
	}

	// Student's own list includes it too.
	w = jsonReq(t, sr, http.MethodGet, "/api/student/submissions", nil)
	var mine []model.Submission
	json.Unmarshal(w.Body.Bytes(), &mine)
	if len(mine) != 1 {
		t.Fatalf("student list should have 1, got %+v", mine)
	}
}

func TestSubmissionPostRejectsUnsupportedHost(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	schoolID := schoolIDForOrg(t, pool, org)
	dj := seedDJ(t, pool, org, "poster2")
	chal := seedCoreChallenge(t, pool, schoolID, org, "Genre Bender")

	sr := submissionStudentRouter(pool, schoolID, dj)
	w := jsonReq(t, sr, http.MethodPost, "/api/student/submissions", map[string]any{
		"challenge_id": chal, "media_url": "https://vimeo.com/12345",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unsupported host: want 400, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestSubmissionReviewNeedsWorkKeepsNoBadge(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	schoolID := schoolIDForOrg(t, pool, org)
	dj := seedDJ(t, pool, org, "poster3")
	chal := seedCoreChallenge(t, pool, schoolID, org, "Crowd Reader")

	sr := submissionStudentRouter(pool, schoolID, dj)
	w := jsonReq(t, sr, http.MethodPost, "/api/student/submissions", map[string]any{
		"challenge_id": chal, "media_url": "https://soundcloud.com/artist/mix",
	})
	var sub model.Submission
	json.Unmarshal(w.Body.Bytes(), &sub)

	tr := submissionTeacherRouter(pool, org)
	w = jsonReq(t, tr, http.MethodPost, "/api/submissions/"+sub.ID+"/review", map[string]any{
		"status": "needs_work", "review_note": "tighten the transition",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("needs_work review: want 200, got %d", w.Code)
	}
	var reviewed model.Submission
	json.Unmarshal(w.Body.Bytes(), &reviewed)
	if reviewed.Status != "needs_work" || reviewed.EarnedBadgeLabel != "" {
		t.Fatalf("needs_work must not award a badge: %+v", reviewed)
	}
}

func TestSubmissionPendingReviewQueue(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	schoolID := schoolIDForOrg(t, pool, org)
	dj := seedDJ(t, pool, org, "poster4")
	chal := seedCoreChallenge(t, pool, schoolID, org, "Headliner")

	sr := submissionStudentRouter(pool, schoolID, dj)
	w := jsonReq(t, sr, http.MethodPost, "/api/student/submissions", map[string]any{
		"challenge_id": chal, "media_url": "https://youtu.be/pending01",
	})
	var sub model.Submission
	json.Unmarshal(w.Body.Bytes(), &sub)

	tr := submissionTeacherRouter(pool, org)
	// Posted submission shows in the review queue.
	w = jsonReq(t, tr, http.MethodGet, "/api/submissions/pending", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("pending: want 200, got %d", w.Code)
	}
	var pending []model.Submission
	json.Unmarshal(w.Body.Bytes(), &pending)
	if len(pending) != 1 || pending[0].ID != sub.ID {
		t.Fatalf("posted submission should be in the review queue, got %+v", pending)
	}

	// After a pass it leaves the queue.
	jsonReq(t, tr, http.MethodPost, "/api/submissions/"+sub.ID+"/review", map[string]any{"status": "passed"})
	w = jsonReq(t, tr, http.MethodGet, "/api/submissions/pending", nil)
	json.Unmarshal(w.Body.Bytes(), &pending)
	if len(pending) != 0 {
		t.Fatalf("passed submission should leave the review queue, got %+v", pending)
	}
}
