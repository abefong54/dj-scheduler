package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
	linenotifyuc "eventlineup/internal/usecase/linenotify"
)

// LineHandler manages the per-event LINE Notify token (US-006).
type LineHandler struct{ uc *linenotifyuc.UseCase }

func NewLineHandler(uc *linenotifyuc.UseCase) *LineHandler { return &LineHandler{uc: uc} }

// Register mounts the LINE token route under the (authenticated) API group.
func (h *LineHandler) Register(rg *gin.RouterGroup) {
	rg.PUT("/events/:id/line-token", h.setToken)
}

// @Summary     Set or clear an event's LINE Notify token
// @Description Stores the token encrypted at rest, or clears it when empty. The raw token is never returned.
// @Tags        events
// @Accept      json
// @Produce     json
// @Param       id    path      string  true  "Event ID (UUID)"
// @Param       body  body      object  true  "{ line_notify_token }"
// @Success     200   {object}  map[string]bool
// @Failure     400   {object}  map[string]string
// @Failure     403   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/events/{id}/line-token [put]
func (h *LineHandler) setToken(c *gin.Context) {
	var body struct {
		LineNotifyToken string `json:"line_notify_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	enabled, err := h.uc.SetToken(c.Request.Context(), c.Param("id"), organizerID, body.LineNotifyToken)
	if errors.Is(err, apperrors.ErrNotFound) {
		// The event isn't the organizer's. Its UUID is already public (shared
		// schedule link), so 403 here leaks nothing GET /:id/public doesn't.
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"line_notify_enabled": enabled})
}
