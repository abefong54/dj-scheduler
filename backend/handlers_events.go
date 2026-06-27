// backend/handlers_events.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerEventRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/events", listEvents(pool))
	rg.POST("/events", createEvent(pool))
	rg.GET("/events/:id", getEvent(pool))
	rg.DELETE("/events/:id", deleteEvent(pool))
	rg.POST("/events/:id/duplicate", duplicateEvent(pool))
}

func listEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT id, name, venue_name, start_date::text, end_date::text
			 FROM events ORDER BY start_date DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		events := []Event{}
		for rows.Next() {
			var e Event
			rows.Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate)
			events = append(events, e)
		}
		c.JSON(http.StatusOK, events)
	}
}

func getEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e Event
		err := pool.QueryRow(context.Background(),
			`SELECT id, name, venue_name, start_date::text, end_date::text
			 FROM events WHERE id = $1`, c.Param("id")).
			Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, e)
	}
}

func createEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e Event
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if e.Name == "" || e.VenueName == "" || e.StartDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name, venue_name, start_date required"})
			return
		}
		if e.EndDate == "" {
			e.EndDate = e.StartDate
		}
		err := pool.QueryRow(context.Background(),
			`INSERT INTO events (name, venue_name, start_date, end_date)
			 VALUES ($1,$2,$3,$4)
			 RETURNING id, name, venue_name, start_date::text, end_date::text`,
			e.Name, e.VenueName, e.StartDate, e.EndDate).
			Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, e)
	}
}

func deleteEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM events WHERE id = $1`, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func duplicateEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		origID := c.Param("id")

		var orig Event
		err := pool.QueryRow(ctx,
			`SELECT name, venue_name, start_date::text, end_date::text
			 FROM events WHERE id = $1`, origID).
			Scan(&orig.Name, &orig.VenueName, &orig.StartDate, &orig.EndDate)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		var newEvent Event
		err = pool.QueryRow(ctx,
			`INSERT INTO events (name, venue_name, start_date, end_date)
			 VALUES ($1||' (copy)', $2, $3, $4)
			 RETURNING id, name, venue_name, start_date::text, end_date::text`,
			orig.Name, orig.VenueName, orig.StartDate, orig.EndDate).
			Scan(&newEvent.ID, &newEvent.Name, &newEvent.VenueName, &newEvent.StartDate, &newEvent.EndDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Copy stages (not slots — dates would need adjustment)
		rows, _ := pool.Query(ctx,
			`SELECT name, color, display_order FROM stages WHERE event_id = $1 ORDER BY display_order`, origID)
		defer rows.Close()
		for rows.Next() {
			var s Stage
			rows.Scan(&s.Name, &s.Color, &s.DisplayOrder)
			pool.Exec(ctx,
				`INSERT INTO stages (event_id, name, color, display_order) VALUES ($1,$2,$3,$4)`,
				newEvent.ID, s.Name, s.Color, s.DisplayOrder)
		}

		c.JSON(http.StatusCreated, newEvent)
	}
}
