package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/interfaces/http/middleware"
	djuc "eventlineup/internal/usecase/dj"
)

type DJHandler struct{ uc *djuc.UseCase }

func NewDJHandler(uc *djuc.UseCase) *DJHandler { return &DJHandler{uc: uc} }

func (h *DJHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/djs", h.list)
	rg.POST("/djs", h.create)
	rg.GET("/djs/:id", h.get)
	rg.PATCH("/djs/:id", h.patch)
	rg.DELETE("/djs/:id", h.delete)
}

// @Summary     List DJs
// @Description Returns all DJs ordered by name
// @Tags        djs
// @Produce     json
// @Success     200  {array}   model.DJ
// @Failure     500  {object}  map[string]string
// @Router      /api/djs [get]
func (h *DJHandler) list(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	djs, err := h.uc.List(c.Request.Context(), organizerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, djs)
}

// @Summary     Get DJ
// @Description Returns a single DJ by ID
// @Tags        djs
// @Produce     json
// @Param       id   path      string   true  "DJ ID (UUID)"
// @Success     200  {object}  model.DJ
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /api/djs/{id} [get]
func (h *DJHandler) get(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	d, err := h.uc.Get(c.Request.Context(), c.Param("id"), organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, map[string]string{"error": "dj not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

// @Summary     Create DJ
// @Description Creates a new DJ
// @Tags        djs
// @Accept      json
// @Produce     json
// @Param       body  body      model.DJ  true  "DJ"
// @Success     201   {object}  model.DJ
// @Failure     400   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/djs [post]
func (h *DJHandler) create(c *gin.Context) {
	var body model.DJ
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	if body.GenreTags == nil {
		body.GenreTags = []string{}
	}
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	d, err := h.uc.Create(c.Request.Context(), body.Name, body.GenreTags, organizerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

// @Summary     Patch DJ
// @Description Updates a DJ's name and genre tags
// @Tags        djs
// @Accept      json
// @Produce     json
// @Param       id    path      string    true  "DJ ID (UUID)"
// @Param       body  body      model.DJ  true  "Fields to update"
// @Success     200   {object}  model.DJ
// @Failure     400   {object}  map[string]string
// @Failure     404   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /api/djs/{id} [patch]
func (h *DJHandler) patch(c *gin.Context) {
	var body model.DJ
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	if body.GenreTags == nil {
		body.GenreTags = []string{}
	}
	body.ID = c.Param("id")
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	d, err := h.uc.Update(c.Request.Context(), body, organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, map[string]string{"error": "dj not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

// @Summary     Delete DJ
// @Description Deletes a DJ by ID
// @Tags        djs
// @Param       id   path  string  true  "DJ ID (UUID)"
// @Success     204
// @Failure     500  {object}  map[string]string
// @Router      /api/djs/{id} [delete]
func (h *DJHandler) delete(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	if err := h.uc.Delete(c.Request.Context(), c.Param("id"), organizerID); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
