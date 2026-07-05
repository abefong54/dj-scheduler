package handler

import (
	"github.com/gin-gonic/gin"

	submissionuc "eventlineup/internal/usecase/submission"
)

// SubmissionHandler serves the student attempt path (post, my-list, feed) on the
// student-authed group and teacher review on the organizer-authed group.
// SCAFFOLD: routes are wired; every method returns 501 until Track C (submission
// backend) fills them. notImplemented is shared with challenge_handler.go.
type SubmissionHandler struct {
	uc *submissionuc.UseCase
}

func NewSubmissionHandler(uc *submissionuc.UseCase) *SubmissionHandler {
	return &SubmissionHandler{uc: uc}
}

// RegisterProtected mounts the teacher review route on the organizer-authed group.
func (h *SubmissionHandler) RegisterProtected(rg *gin.RouterGroup) {
	rg.POST("/submissions/:id/review", h.review)
}

// RegisterStudent mounts the student routes on the student-authed group.
func (h *SubmissionHandler) RegisterStudent(rg *gin.RouterGroup) {
	rg.POST("/student/submissions", h.postAttempt)
	rg.GET("/student/submissions", h.listForStudent)
	rg.GET("/student/feed", h.feed)
}

// --- teacher (organizer JWT + school scope) ---

func (h *SubmissionHandler) review(c *gin.Context) { notImplemented(c) }

// --- student (student JWT) ---

func (h *SubmissionHandler) postAttempt(c *gin.Context)    { notImplemented(c) }
func (h *SubmissionHandler) listForStudent(c *gin.Context) { notImplemented(c) }
func (h *SubmissionHandler) feed(c *gin.Context)           { notImplemented(c) }
