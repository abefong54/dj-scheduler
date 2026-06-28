package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	stageuc "eventlineup/internal/usecase/stage"
)

type StageHandler struct{ uc *stageuc.UseCase }

func NewStageHandler(uc *stageuc.UseCase) *StageHandler { return &StageHandler{uc: uc} }

func (h *StageHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/events/:id/stages", h.list)
	rg.POST("/events/:id/stages", h.create)
	rg.GET("/events/:id/stages/:stage_id", h.get)
	rg.PATCH("/events/:id/stages/:stage_id", h.patch)
	rg.DELETE("/events/:id/stages/:stage_id", h.delete)
}

// @Summary     List Stages
// @Description Returns all stages for an event ordered by display_order
// @Tags        stages
// @Produce     json
// @Param       id   path      string  true  "Event ID (UUID)"
// @Success     200  {array}   model.Stage
// @Failure     500  {object}  map[string]string
// @Router      /api/events/{id}/stages [get]
func (h *StageHandler) list(c *gin.Context) {
	stages, err := h.uc.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stages)
}

// @Summary     Get Stage
// @Description Returns a single stage by ID within an event
// @Tags        stages
// @Produce     json
// @Param       id        path      string  true  "Event ID (UUID)"
// @Param       stage_id  path      string  true  "Stage ID (UUID)"
// @Success     200       {object}  model.Stage
// @Failure     404       {object}  map[string]string
// @Failure     500       {object}  map[string]string
// @Router      /api/events/{id}/stages/{stage_id} [get]
func (h *StageHandler) get(c *gin.Context) {
	s, err := h.uc.Get(c.Request.Context(), c.Param("stage_id"), c.Param("id"))
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "stage not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// @Summary     Create Stage
// @Description Creates a new stage within an event
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       id    path      string       true  "Event ID (UUID)"
// @Param       body  body      model.Stage  true  "Stage"
// @Success     201   {object}  model.Stage
// @Failure     400   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/events/{id}/stages [post]
func (h *StageHandler) create(c *gin.Context) {
	var body model.Stage
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	if body.Color == "" {
		body.Color = "#6366F1"
	}
	s, err := h.uc.Create(c.Request.Context(), c.Param("id"), body.Name, body.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// @Summary     Patch Stage
// @Description Updates a stage's name and color
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       id        path      string       true  "Event ID (UUID)"
// @Param       stage_id  path      string       true  "Stage ID (UUID)"
// @Param       body      body      model.Stage  true  "Fields to update"
// @Success     200       {object}  model.Stage
// @Failure     400       {object}  map[string]string
// @Failure     404       {object}  map[string]string
// @Failure     500       {object}  map[string]string
// @Router      /api/events/{id}/stages/{stage_id} [patch]
func (h *StageHandler) patch(c *gin.Context) {
	var body model.Stage
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	if body.Color == "" {
		body.Color = "#6366F1"
	}
	body.ID = c.Param("stage_id")
	body.EventID = c.Param("id")
	s, err := h.uc.Update(c.Request.Context(), body)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "stage not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// @Summary     Delete Stage
// @Description Deletes a stage from an event
// @Tags        stages
// @Param       id        path  string  true  "Event ID (UUID)"
// @Param       stage_id  path  string  true  "Stage ID (UUID)"
// @Success     204
// @Failure     500  {object}  map[string]string
// @Router      /api/events/{id}/stages/{stage_id} [delete]
func (h *StageHandler) delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.Param("stage_id"), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
