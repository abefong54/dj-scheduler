package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/interfaces/http/middleware"
	challengeuc "eventlineup/internal/usecase/challenge"
)

// ChallengeHandler serves teacher challenge CRUD (organizer-authed, school-scoped)
// and the student browse/list surfaces (student-authed).
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

// challengeWriteRequest contains only the fields a teacher may set on create or
// update. type/assignee/badge invariants are enforced in the usecase.
type challengeWriteRequest struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Skill        string `json:"skill"`
	AssignedDJID string `json:"assigned_dj_id"`
	BadgeLabel   string `json:"badge_label"`
	BadgeIcon    string `json:"badge_icon"`
	SortOrder    int    `json:"sort_order"`
}

func (r challengeWriteRequest) toModel() model.Challenge {
	return model.Challenge{
		Type:         r.Type,
		Title:        r.Title,
		Description:  r.Description,
		Skill:        r.Skill,
		AssignedDJID: r.AssignedDJID,
		BadgeLabel:   r.BadgeLabel,
		BadgeIcon:    r.BadgeIcon,
		SortOrder:    r.SortOrder,
	}
}

// --- teacher (organizer JWT + school scope) ---

func (h *ChallengeHandler) list(c *gin.Context) {
	schoolID := c.GetString(middleware.SchoolIDKey)
	challenges, err := h.uc.List(c.Request.Context(), schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, challenges)
}

func (h *ChallengeHandler) create(c *gin.Context) {
	var req challengeWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	m := req.toModel()
	m.CreatedBy = c.GetString(middleware.OrganizerIDKey)
	schoolID := c.GetString(middleware.SchoolIDKey)
	created, err := h.uc.Create(c.Request.Context(), m, schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ChallengeHandler) update(c *gin.Context) {
	var req challengeWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	m := req.toModel()
	m.ID = c.Param("id")
	schoolID := c.GetString(middleware.SchoolIDKey)
	updated, err := h.uc.Update(c.Request.Context(), m, schoolID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ChallengeHandler) archive(c *gin.Context) {
	schoolID := c.GetString(middleware.SchoolIDKey)
	err := h.uc.Archive(c.Request.Context(), c.Param("id"), schoolID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- student (student JWT) ---

func (h *ChallengeHandler) listForStudent(c *gin.Context) {
	djID := c.GetString(middleware.StudentIDKey)
	schoolID := c.GetString(middleware.SchoolIDKey)
	challenges, err := h.uc.ListForStudent(c.Request.Context(), djID, schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, challenges)
}

func (h *ChallengeHandler) browseCore(c *gin.Context) {
	schoolID := c.GetString(middleware.SchoolIDKey)
	challenges, err := h.uc.ListCore(c.Request.Context(), schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, challenges)
}

// notImplemented is the shared scaffold response: an obvious 501 so downstream
// agents know the route is wired but the logic is theirs to fill. Still used by
// submission_handler.go (Track C) until those routes are filled.
func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
