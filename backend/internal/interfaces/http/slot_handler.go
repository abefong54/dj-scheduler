package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/interfaces/http/middleware"
	slotuc "eventlineup/internal/usecase/slot"
)

type SlotHandler struct{ uc *slotuc.UseCase }

func NewSlotHandler(uc *slotuc.UseCase) *SlotHandler { return &SlotHandler{uc: uc} }

func (h *SlotHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/events/:id/slots", h.list)
	rg.POST("/events/:id/slots", h.create)
	rg.GET("/events/:id/slots/:slot_id", h.get)
	rg.PATCH("/events/:id/slots/:slot_id", h.update)
	rg.DELETE("/events/:id/slots/:slot_id", h.delete)
}

// slotWriteRequest contains only the fields a client may set on create or update.
type slotWriteRequest struct {
	StageID   string `json:"stage_id"`
	DjID      string `json:"dj_id"`
	Genre     string `json:"genre"`
	SlotDate  string `json:"slot_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Notes     string `json:"notes"`
}

func (r slotWriteRequest) validate() bool {
	return r.StageID != "" && r.SlotDate != "" && r.StartTime != "" && r.EndTime != ""
}

// writeSlotError maps a usecase error to its HTTP response: 404 for a missing
// slot, 409 with structured details for a scheduling conflict, 500 otherwise.
func writeSlotError(c *gin.Context, err error) {
	var conflict *apperrors.ConflictError
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "slot not found"})
	case errors.As(err, &conflict):
		c.JSON(http.StatusConflict, gin.H{
			"error":               "conflict",
			"type":                conflict.Type,
			"conflicting_slot_id": conflict.ConflictingSlotID,
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// @Summary     List Slots
// @Description Returns all slots for an event ordered by date and start time
// @Tags        slots
// @Produce     json
// @Param       id   path      string  true  "Event ID (UUID)"
// @Success     200  {array}   model.Slot
// @Failure     500  {object}  map[string]string
// @Router      /api/events/{id}/slots [get]
func (h *SlotHandler) list(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	slots, err := h.uc.List(c.Request.Context(), c.Param("id"), organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, slots)
}

// @Summary     Get Slot
// @Description Returns a single slot by ID within an event
// @Tags        slots
// @Produce     json
// @Param       id       path      string  true  "Event ID (UUID)"
// @Param       slot_id  path      string  true  "Slot ID (UUID)"
// @Success     200      {object}  model.Slot
// @Failure     404      {object}  map[string]string
// @Failure     500      {object}  map[string]string
// @Router      /api/events/{id}/slots/{slot_id} [get]
func (h *SlotHandler) get(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	s, err := h.uc.Get(c.Request.Context(), c.Param("slot_id"), c.Param("id"), organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "slot not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// @Summary     Create Slot
// @Description Creates a new slot within an event
// @Tags        slots
// @Accept      json
// @Produce     json
// @Param       id    path      string            true  "Event ID (UUID)"
// @Param       body  body      slotWriteRequest  true  "Slot"
// @Success     201   {object}  model.Slot
// @Failure     400   {object}  map[string]string
// @Failure     409   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/events/{id}/slots [post]
func (h *SlotHandler) create(c *gin.Context) {
	var req slotWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if !req.validate() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stage_id, slot_date, start_time, end_time required"})
		return
	}

	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	s, err := h.uc.Create(c.Request.Context(), model.Slot{
		StageID:   req.StageID,
		DjID:      req.DjID,
		Genre:     req.Genre,
		SlotDate:  req.SlotDate,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Notes:     req.Notes,
	}, c.Param("id"), organizerID)
	if err != nil {
		writeSlotError(c, err)
		return
	}
	c.JSON(http.StatusCreated, s)
}

// @Summary     Patch Slot
// @Description Updates a slot's fields within an event
// @Tags        slots
// @Accept      json
// @Produce     json
// @Param       id       path      string            true  "Event ID (UUID)"
// @Param       slot_id  path      string            true  "Slot ID (UUID)"
// @Param       body     body      slotWriteRequest  true  "Fields to update"
// @Success     200      {object}  model.Slot
// @Failure     400      {object}  map[string]string
// @Failure     404      {object}  map[string]string
// @Failure     409      {object}  map[string]string
// @Failure     500      {object}  map[string]string
// @Router      /api/events/{id}/slots/{slot_id} [patch]
func (h *SlotHandler) update(c *gin.Context) {
	var req slotWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if !req.validate() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stage_id, slot_date, start_time, end_time required"})
		return
	}

	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	s, err := h.uc.Update(c.Request.Context(), model.Slot{
		ID:        c.Param("slot_id"),
		StageID:   req.StageID,
		DjID:      req.DjID,
		Genre:     req.Genre,
		SlotDate:  req.SlotDate,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Notes:     req.Notes,
	}, c.Param("id"), organizerID)
	if err != nil {
		writeSlotError(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}

// @Summary     Delete Slot
// @Description Deletes a slot from an event
// @Tags        slots
// @Param       id       path  string  true  "Event ID (UUID)"
// @Param       slot_id  path  string  true  "Slot ID (UUID)"
// @Success     204
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /api/events/{id}/slots/{slot_id} [delete]
func (h *SlotHandler) delete(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	err := h.uc.Delete(c.Request.Context(), c.Param("slot_id"), c.Param("id"), organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "slot not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
