package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
	djuc "eventlineup/internal/usecase/dj"
	slotuc "eventlineup/internal/usecase/slot"
)

// bearerPrefix is the scheme prefix on the Authorization header.
const bearerPrefix = "Bearer "

// portalToken extracts the raw DJ portal token from the Authorization header.
// EL-038: the token travels in a header, never the URL, so it can't leak via
// browser history, Referer, or access logs. Returns "" when absent/malformed.
func portalToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))
}

// DJPortalHandler serves the DJ self-service portal (US-009): organizers mint
// per-DJ tokens on an authenticated route, and DJs redeem those tokens on a
// public route (the token itself is the credential — no login).
type DJPortalHandler struct {
	uc          *djuc.UseCase
	slots       *slotuc.UseCase
	frontendURL string
}

func NewDJPortalHandler(uc *djuc.UseCase, slots *slotuc.UseCase, frontendURL string) *DJPortalHandler {
	return &DJPortalHandler{uc: uc, slots: slots, frontendURL: frontendURL}
}

// RegisterProtected registers routes that require an organizer JWT.
func (h *DJPortalHandler) RegisterProtected(rg *gin.RouterGroup) {
	rg.POST("/djs/:id/token", h.generateToken)
}

// RegisterPublic registers routes that are reachable without auth (token-gated).
func (h *DJPortalHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.GET("/dj/portal", h.portal)
	rg.PATCH("/dj/portal/slots/:slot_id", h.confirmSlot)
}

// @Summary     Generate DJ portal token
// @Description Issues a personal, expiring portal link for a DJ. Regenerating revokes the previous token.
// @Tags        djs
// @Produce     json
// @Param       id   path      string  true  "DJ ID (UUID)"
// @Success     200  {object}  map[string]interface{}
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /api/djs/{id}/token [post]
func (h *DJPortalHandler) generateToken(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	token, expiresAt, err := h.uc.GeneratePortalToken(c.Request.Context(), c.Param("id"), organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		// 404 (not 403) for a DJ the organizer doesn't own — consistent with
		// US-002 scoping, which avoids leaking the existence of others' rows.
		c.JSON(http.StatusNotFound, gin.H{"error": "dj not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"portal_url": h.frontendURL + "/dj/portal?token=" + token,
		"expires_at": expiresAt,
	})
}

// @Summary     DJ portal view
// @Description Returns the DJ and their slots across all events for a valid portal token. No auth required.
// @Tags        public
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer <raw portal token>"
// @Success     200    {object}  map[string]interface{}
// @Failure     401    {object}  map[string]string
// @Router      /api/dj/portal [get]
func (h *DJPortalHandler) portal(c *gin.Context) {
	rawToken := portalToken(c)
	if rawToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}
	d, slots, err := h.uc.PortalByToken(c.Request.Context(), rawToken)
	if err != nil {
		// Unknown, tampered, or expired token — all surface as 401 without detail.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dj": d, "slots": slots})
}

// @Summary     DJ confirms or flags a slot
// @Description Sets the DJ's response on one of their slots. Token-gated, no organizer auth.
// @Tags        public
// @Accept      json
// @Produce     json
// @Param       slot_id  path      string  true  "Slot ID (UUID)"
// @Param       Authorization  header  string  true  "Bearer <raw portal token>"
// @Success     200      {object}  map[string]interface{}
// @Failure     400      {object}  map[string]string
// @Failure     401      {object}  map[string]string
// @Failure     403      {object}  map[string]string
// @Router      /api/dj/portal/slots/{slot_id} [patch]
func (h *DJPortalHandler) confirmSlot(c *gin.Context) {
	rawToken := portalToken(c)
	if rawToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}
	var body struct {
		Confirmation string `json:"confirmation"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Confirmation != "confirmed" && body.Confirmation != "flagged" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confirmation must be 'confirmed' or 'flagged'"})
		return
	}

	dj, err := h.uc.DJByToken(c.Request.Context(), rawToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	err = h.slots.SetDJConfirmation(c.Request.Context(), c.Param("slot_id"), dj.ID, body.Confirmation)
	if errors.Is(err, apperrors.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": "slot does not belong to this DJ"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": c.Param("slot_id"), "dj_confirmation": body.Confirmation})
}
