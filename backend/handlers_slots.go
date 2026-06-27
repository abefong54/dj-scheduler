// backend/handlers_slots.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerSlotRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/events/:id/slots", listSlots(pool))
	rg.POST("/events/:id/slots", createSlot(pool))
	rg.DELETE("/events/:id/slots/:slot_id", deleteSlot(pool))
}

func listSlots(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(), `
			SELECT sl.id, sl.event_id, sl.stage_id, st.name,
			       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
			       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,'')
			FROM slots sl
			JOIN stages st ON st.id = sl.stage_id
			LEFT JOIN djs d ON d.id = sl.dj_id
			WHERE sl.event_id = $1
			ORDER BY sl.slot_date, sl.start_time`, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		slots := []Slot{}
		for rows.Next() {
			var s Slot
			rows.Scan(&s.ID, &s.EventID, &s.StageID, &s.StageName,
				&s.DjID, &s.DjName, &s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes)
			slots = append(slots, s)
		}
		c.JSON(http.StatusOK, slots)
	}
}

func createSlot(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s Slot
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if s.StageID == "" || s.SlotDate == "" || s.StartTime == "" || s.EndTime == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage_id, slot_date, start_time, end_time required"})
			return
		}
		eventID := c.Param("id")
		// dj_id is optional — store NULL when empty
		err := pool.QueryRow(context.Background(),
			`INSERT INTO slots (event_id, stage_id, dj_id, slot_date, start_time, end_time, notes)
			 VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7)
			 RETURNING id`,
			eventID, s.StageID, s.DjID, s.SlotDate, s.StartTime, s.EndTime, s.Notes).
			Scan(&s.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		s.EventID = eventID
		c.JSON(http.StatusCreated, s)
	}
}

func deleteSlot(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM slots WHERE id = $1 AND event_id = $2`,
			c.Param("slot_id"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
