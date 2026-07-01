package handler

import (
	"net/http"
	"net/mail"
	"strings"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	leaduc "eventlineup/internal/usecase/lead"

	"github.com/gin-gonic/gin"
)

type LeadHandler struct{ uc *leaduc.UseCase }

func NewLeadHandler(uc *leaduc.UseCase) *LeadHandler { return &LeadHandler{uc: uc} }

func (h *LeadHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/leads", h.create)
}

type leadWriteRequest struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Email        string `json:"email"`
	Message      string `json:"message"`
}

const (
	maxLeadName  = 200
	maxLeadOrg   = 200
	maxLeadEmail = 320
	maxLeadMsg   = 2000
)

func (r leadWriteRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	email := strings.TrimSpace(r.Email)
	if name == "" || email == "" {
		return apperrors.NewValidation("name and email are required")
	}
	if len(name) > maxLeadName || len(r.Organization) > maxLeadOrg ||
		len(email) > maxLeadEmail || len(r.Message) > maxLeadMsg {
		return apperrors.NewValidation("one or more fields exceed the maximum length")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return apperrors.NewValidation("a valid email is required")
	}
	return nil
}

// create is PUBLIC — no organizer in context, so never read OrganizerIDKey.
func (h *LeadHandler) create(c *gin.Context) {
	var body leadWriteRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := body.validate(); err != nil {
		respondError(c, err) // EL-039-safe; never logs the payload
		return
	}
	created, err := h.uc.Create(c.Request.Context(), model.Lead{
		Name:         strings.TrimSpace(body.Name),
		Organization: strings.TrimSpace(body.Organization),
		Email:        strings.TrimSpace(body.Email),
		Message:      strings.TrimSpace(body.Message),
		Source:       "landing",
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}
