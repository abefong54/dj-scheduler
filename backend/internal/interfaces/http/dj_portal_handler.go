package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
	djuc "eventlineup/internal/usecase/dj"
)

// DJPortalHandler serves the DJ self-service portal (US-009): organizers mint
// per-DJ tokens on an authenticated route, and DJs redeem those tokens on a
// public route (the token itself is the credential — no login).
type DJPortalHandler struct {
	uc          *djuc.UseCase
	frontendURL string
}

func NewDJPortalHandler(uc *djuc.UseCase, frontendURL string) *DJPortalHandler {
	return &DJPortalHandler{uc: uc, frontendURL: frontendURL}
}

// RegisterProtected registers routes that require an organizer JWT.
func (h *DJPortalHandler) RegisterProtected(rg *gin.RouterGroup) {
	rg.POST("/djs/:id/token", h.generateToken)
}

// RegisterPublic registers routes that are reachable without auth (token-gated).
func (h *DJPortalHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.GET("/dj/portal", h.portal)
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
// @Param       token  query     string  true  "Raw portal token"
// @Success     200    {object}  map[string]interface{}
// @Failure     401    {object}  map[string]string
// @Router      /api/dj/portal [get]
func (h *DJPortalHandler) portal(c *gin.Context) {
	rawToken := c.Query("token")
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
