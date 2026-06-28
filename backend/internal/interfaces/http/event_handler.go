package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/interfaces/http/middleware"
	eventuc "eventlineup/internal/usecase/event"
)

type EventHandler struct{ uc *eventuc.UseCase }

func NewEventHandler(uc *eventuc.UseCase) *EventHandler { return &EventHandler{uc: uc} }

func (h *EventHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/events", h.list)
	rg.POST("/events", h.create)
	rg.GET("/events/:id", h.get)
	rg.PATCH("/events/:id", h.patch)
	rg.DELETE("/events/:id", h.delete)
	rg.POST("/events/:id/duplicate", h.duplicate)
}

// @Summary     List Events
// @Description Returns all events ordered by start date descending
// @Tags        events
// @Produce     json
// @Success     200  {array}   model.Event
// @Failure     500  {object}  map[string]string
// @Router      /api/events [get]
func (h *EventHandler) list(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	events, err := h.uc.List(c.Request.Context(), organizerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// @Summary     Get Event
// @Description Returns a single event by ID
// @Tags        events
// @Produce     json
// @Param       id   path      string  true  "Event ID (UUID)"
// @Success     200  {object}  model.Event
// @Failure     404  {object}  map[string]string
// @Router      /api/events/{id} [get]
func (h *EventHandler) get(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	e, err := h.uc.Get(c.Request.Context(), c.Param("id"), organizerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, e)
}

// @Summary     Create Event
// @Description Creates a new event
// @Tags        events
// @Accept      json
// @Produce     json
// @Param       body  body      model.Event  true  "Event"
// @Success     201   {object}  model.Event
// @Failure     400   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/events [post]
func (h *EventHandler) create(c *gin.Context) {
	var body model.Event
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Name == "" || body.VenueName == "" || body.StartDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, venue_name, start_date required"})
		return
	}
	if body.EndDate == "" {
		body.EndDate = body.StartDate
	}
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	e, err := h.uc.Create(c.Request.Context(), body, organizerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

// @Summary     Patch Event
// @Description Updates an event's name, venue, dates, and genres
// @Tags        events
// @Accept      json
// @Produce     json
// @Param       id    path      string       true  "Event ID (UUID)"
// @Param       body  body      model.Event  true  "Fields to update"
// @Success     200   {object}  model.Event
// @Failure     400   {object}  map[string]string
// @Failure     404   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/events/{id} [patch]
func (h *EventHandler) patch(c *gin.Context) {
	var body model.Event
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Name == "" || body.VenueName == "" || body.StartDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, venue_name, start_date required"})
		return
	}
	if body.EndDate == "" {
		body.EndDate = body.StartDate
	}
	if body.Genres == nil {
		body.Genres = []string{}
	}
	body.ID = c.Param("id")
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	e, err := h.uc.Update(c.Request.Context(), body, organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, e)
}

// @Summary     Delete Event
// @Description Deletes an event and its stages/slots (cascade)
// @Tags        events
// @Param       id   path  string  true  "Event ID (UUID)"
// @Success     204
// @Failure     500  {object}  map[string]string
// @Router      /api/events/{id} [delete]
func (h *EventHandler) delete(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	if err := h.uc.Delete(c.Request.Context(), c.Param("id"), organizerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Duplicate Event
// @Description Creates a copy of an event including its stages
// @Tags        events
// @Produce     json
// @Param       id   path      string  true  "Event ID (UUID)"
// @Success     201  {object}  model.Event
// @Failure     404  {object}  map[string]string
// @Router      /api/events/{id}/duplicate [post]
func (h *EventHandler) duplicate(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	e, err := h.uc.Duplicate(c.Request.Context(), c.Param("id"), organizerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusCreated, e)
}
