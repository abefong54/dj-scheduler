package handler

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
	schooluc "eventlineup/internal/usecase/school"
)

type SchoolHandler struct{ uc *schooluc.UseCase }

func NewSchoolHandler(uc *schooluc.UseCase) *SchoolHandler { return &SchoolHandler{uc: uc} }

// Register mounts the onboarding surface (EL-105) on the authenticated group.
func (h *SchoolHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/me", h.me)
	rg.POST("/schools", h.register)
	rg.GET("/schools/current/members", h.listMembers)
	rg.POST("/schools/current/members", h.invite)
}

type userDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type schoolDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}
type meResponse struct {
	User   userDTO    `json:"user"`
	School *schoolDTO `json:"school"`
}

// me returns the caller's user record and their school (null before onboarding).
func (h *SchoolHandler) me(c *gin.Context) {
	organizerID := c.GetString(middleware.OrganizerIDKey)
	org, sm, err := h.uc.Me(c.Request.Context(), organizerID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := meResponse{User: userDTO{ID: org.ID, Name: org.Name, Email: org.Email}}
	if sm != nil {
		resp.School = &schoolDTO{ID: sm.School.ID, Name: sm.School.Name, Role: sm.Role}
	}
	c.JSON(http.StatusOK, resp)
}

type registerSchoolRequest struct {
	Name string `json:"name"`
	City string `json:"city"`
}

func (h *SchoolHandler) register(c *gin.Context) {
	var body registerSchoolRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		respondError(c, apperrors.NewValidation("school name is required"))
		return
	}
	if len(name) > 200 || len(body.City) > 200 {
		respondError(c, apperrors.NewValidation("name or city exceeds the maximum length"))
		return
	}
	organizerID := c.GetString(middleware.OrganizerIDKey)
	school, role, err := h.uc.Register(c.Request.Context(), organizerID, name, strings.TrimSpace(body.City))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": school.ID, "name": school.Name, "city": school.City, "role": role})
}

func (h *SchoolHandler) listMembers(c *gin.Context) {
	schoolID := c.GetString(middleware.SchoolIDKey)
	members, err := h.uc.ListMembers(c.Request.Context(), schoolID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// invite adds a pending membership. Owner/admin only (the role comes from the
// School middleware, so a member or a school-less caller is 403).
func (h *SchoolHandler) invite(c *gin.Context) {
	if role := c.GetString(middleware.RoleKey); role != "owner" && role != "admin" {
		respondError(c, apperrors.ErrForbidden)
		return
	}
	var body inviteRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	email := strings.TrimSpace(body.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		respondError(c, apperrors.NewValidation("a valid email is required"))
		return
	}
	inviteRole := strings.TrimSpace(body.Role)
	if inviteRole == "" {
		inviteRole = "member"
	}
	m, err := h.uc.Invite(c.Request.Context(), c.GetString(middleware.SchoolIDKey), email, inviteRole)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}
