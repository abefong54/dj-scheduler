package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/interfaces/http/middleware"
	submissionuc "eventlineup/internal/usecase/submission"
)

// SubmissionHandler serves the student attempt path (post, my-list, feed) on the
// student-authed group and teacher review on the organizer-authed group.
type SubmissionHandler struct {
	uc *submissionuc.UseCase
}

func NewSubmissionHandler(uc *submissionuc.UseCase) *SubmissionHandler {
	return &SubmissionHandler{uc: uc}
}

// RegisterProtected mounts the teacher routes on the organizer-authed group.
func (h *SubmissionHandler) RegisterProtected(rg *gin.RouterGroup) {
	rg.GET("/submissions/pending", h.pending)
	rg.POST("/submissions/:id/review", h.review)
}

// RegisterStudent mounts the student routes on the student-authed group.
func (h *SubmissionHandler) RegisterStudent(rg *gin.RouterGroup) {
	rg.POST("/student/submissions", h.postAttempt)
	rg.GET("/student/submissions", h.listForStudent)
	rg.GET("/student/feed", h.feed)
}

type postAttemptRequest struct {
	ChallengeID string `json:"challenge_id"`
	MediaURL    string `json:"media_url"`
	Caption     string `json:"caption"`
}

type reviewRequest struct {
	Status     string `json:"status"` // "passed" | "needs_work"
	ReviewNote string `json:"review_note"`
}

// --- teacher (organizer JWT + school scope) ---

func (h *SubmissionHandler) pending(c *gin.Context) {
	schoolID := c.GetString(middleware.SchoolIDKey)
	subs, err := h.uc.Pending(c.Request.Context(), schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, subs)
}

func (h *SubmissionHandler) review(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	schoolID := c.GetString(middleware.SchoolIDKey)
	updated, err := h.uc.Review(c.Request.Context(), c.Param("id"), schoolID, req.Status, req.ReviewNote)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// --- student (student JWT) ---

func (h *SubmissionHandler) postAttempt(c *gin.Context) {
	var req postAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	djID := c.GetString(middleware.StudentIDKey)
	schoolID := c.GetString(middleware.SchoolIDKey)
	created, err := h.uc.Post(c.Request.Context(), djID, schoolID, req.ChallengeID, req.MediaURL, req.Caption)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *SubmissionHandler) listForStudent(c *gin.Context) {
	djID := c.GetString(middleware.StudentIDKey)
	schoolID := c.GetString(middleware.SchoolIDKey)
	subs, err := h.uc.ListForStudent(c.Request.Context(), djID, schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, subs)
}

func (h *SubmissionHandler) feed(c *gin.Context) {
	schoolID := c.GetString(middleware.SchoolIDKey)
	subs, err := h.uc.Feed(c.Request.Context(), schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, subs)
}
