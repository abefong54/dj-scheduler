package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	challengeuc "eventlineup/internal/usecase/challenge"
)

// ChallengeHandler serves teacher challenge CRUD (organizer-authed, school-scoped)
// and the student browse/list surfaces (student-authed). SCAFFOLD: routes are
// wired; every method returns 501 until Track B (challenge backend) fills them.
type ChallengeHandler struct {
	uc *challengeuc.UseCase
}

func NewChallengeHandler(uc *challengeuc.UseCase) *ChallengeHandler {
	return &ChallengeHandler{uc: uc}
}

// RegisterProtected mounts the teacher routes on the organizer-authed group.
func (h *ChallengeHandler) RegisterProtected(rg *gin.RouterGroup) {
	rg.GET("/challenges", h.list)
	rg.POST("/challenges", h.create)
	rg.PUT("/challenges/:id", h.update)
	rg.DELETE("/challenges/:id", h.archive)
}

// RegisterStudent mounts the student browse/list routes on the student-authed group.
func (h *ChallengeHandler) RegisterStudent(rg *gin.RouterGroup) {
	rg.GET("/student/challenges", h.listForStudent)
	rg.GET("/student/challenges/core", h.browseCore)
}

// --- teacher (organizer JWT + school scope) ---

func (h *ChallengeHandler) list(c *gin.Context)    { notImplemented(c) }
func (h *ChallengeHandler) create(c *gin.Context)  { notImplemented(c) }
func (h *ChallengeHandler) update(c *gin.Context)  { notImplemented(c) }
func (h *ChallengeHandler) archive(c *gin.Context) { notImplemented(c) }

// --- student (student JWT) ---

func (h *ChallengeHandler) listForStudent(c *gin.Context) { notImplemented(c) }
func (h *ChallengeHandler) browseCore(c *gin.Context)     { notImplemented(c) }

// notImplemented is the shared scaffold response: an obvious 501 so downstream
// agents know the route is wired but the logic is theirs to fill.
func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
