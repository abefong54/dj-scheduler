package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/repository"
	"eventlineup/internal/interfaces/http/middleware"
	perfuc "eventlineup/internal/usecase/performance"
)

type PerformanceHandler struct{ uc *perfuc.UseCase }

func NewPerformanceHandler(uc *perfuc.UseCase) *PerformanceHandler {
	return &PerformanceHandler{uc: uc}
}

func (h *PerformanceHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/djs/:id/performance", h.djPerformance)
	rg.GET("/performance", h.roster)
	rg.GET("/performance/underserved", h.underserved)
}

// queryFilter builds a PerformanceFilter from the optional event_id/from/to query
// params shared by the roster and under-served endpoints.
func queryFilter(c *gin.Context) repository.PerformanceFilter {
	return repository.PerformanceFilter{
		EventID: c.Query("event_id"),
		From:    c.Query("from"),
		To:      c.Query("to"),
	}
}

// @Summary     DJ performance
// @Description One DJ's reps across the organizer's events: count, total minutes, last-played date, and a by-genre breakdown (EL-043).
// @Tags        performance
// @Produce     json
// @Param       id   path      string  true  "DJ ID (UUID)"
// @Success     200  {object}  model.DJPerformance
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /api/djs/{id}/performance [get]
func (h *PerformanceHandler) djPerformance(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	p, err := h.uc.DJPerformance(c.Request.Context(), c.Param("id"), organizerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, map[string]string{"error": "dj not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// @Summary     Roster performance summary
// @Description Per-student rep count and total stage time for every active student, including those with zero reps. Optional filters: event_id, from, to (YYYY-MM-DD). EL-043.
// @Tags        performance
// @Produce     json
// @Param       event_id  query  string  false  "Restrict to one event"
// @Param       from      query  string  false  "Inclusive slot_date lower bound (YYYY-MM-DD)"
// @Param       to        query  string  false  "Inclusive slot_date upper bound (YYYY-MM-DD)"
// @Success     200  {array}   model.RosterPerformance
// @Failure     500  {object}  map[string]string
// @Router      /api/performance [get]
func (h *PerformanceHandler) roster(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	rows, err := h.uc.RosterSummary(c.Request.Context(), organizerID, queryFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// @Summary     Under-served students
// @Description Active students below the rep threshold in the window (graduates excluded). threshold defaults when omitted/invalid. EL-043.
// @Tags        performance
// @Produce     json
// @Param       threshold  query  int     false  "Reps below this count as under-served (default applied when omitted)"
// @Param       event_id   query  string  false  "Restrict to one event"
// @Param       from       query  string  false  "Inclusive slot_date lower bound (YYYY-MM-DD)"
// @Param       to         query  string  false  "Inclusive slot_date upper bound (YYYY-MM-DD)"
// @Success     200  {array}   model.RosterPerformance
// @Failure     500  {object}  map[string]string
// @Router      /api/performance/underserved [get]
func (h *PerformanceHandler) underserved(c *gin.Context) {
	organizerID := c.MustGet(middleware.OrganizerIDKey).(string)
	// An absent or unparseable threshold yields 0, which the usecase maps to its
	// default — so a bad value degrades to the default rather than erroring.
	threshold, _ := strconv.Atoi(c.Query("threshold"))
	rows, err := h.uc.Underserved(c.Request.Context(), organizerID, queryFilter(c), threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}
