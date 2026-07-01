package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	eventuc "eventlineup/internal/usecase/event"
	slotuc "eventlineup/internal/usecase/slot"
	stageuc "eventlineup/internal/usecase/stage"
)

// PublicHandler serves unauthenticated, read-only views of an event's schedule.
// It is the bypass target for the auth middleware so attendees can open the
// shared schedule link without signing in.
type PublicHandler struct {
	events *eventuc.UseCase
	stages *stageuc.UseCase
	slots  *slotuc.UseCase
}

func NewPublicHandler(events *eventuc.UseCase, stages *stageuc.UseCase, slots *slotuc.UseCase) *PublicHandler {
	return &PublicHandler{events: events, stages: stages, slots: slots}
}

func (h *PublicHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/events/:id/public", h.get)
	rg.GET("/slots/:id/public", h.getSlot)
}

// @Summary     Public Event Schedule
// @Description Returns an event with its stages and slots in a single payload. No auth required.
// @Tags        public
// @Produce     json
// @Param       id   path      string  true  "Event ID (UUID)"
// @Success     200  {object}  map[string]interface{}
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /api/events/{id}/public [get]
func (h *PublicHandler) get(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	event, err := h.events.GetPublic(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	stages, err := h.stages.ListPublic(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	slots, err := h.slots.ListPublic(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event":  event,
		"stages": stages,
		"slots":  slots,
	})
}

// @Summary     Public Slot
// @Description Returns a single slot with its parent event, by slot ID. No auth required. Backs the per-DJ share card (EL-049).
// @Tags        public
// @Produce     json
// @Param       id   path      string  true  "Slot ID (UUID)"
// @Success     200  {object}  map[string]interface{}
// @Failure     404  {object}  map[string]string
// @Router      /api/slots/{id}/public [get]
func (h *PublicHandler) getSlot(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	slot, err := h.slots.GetPublicByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	event, err := h.events.GetPublic(ctx, slot.EventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slot":  slot,
		"event": event,
	})
}
