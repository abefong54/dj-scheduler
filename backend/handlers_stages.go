// backend/handlers_stages.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerStageRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/events/:id/stages", listStages(pool))
	rg.POST("/events/:id/stages", createStage(pool))
	rg.DELETE("/events/:id/stages/:stage_id", deleteStage(pool))
}

func listStages(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT id, event_id, name, color, display_order
			 FROM stages WHERE event_id = $1 ORDER BY display_order, name`,
			c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		stages := []Stage{}
		for rows.Next() {
			var s Stage
			rows.Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
			stages = append(stages, s)
		}
		c.JSON(http.StatusOK, stages)
	}
}

func createStage(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s Stage
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if s.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
			return
		}
		if s.Color == "" {
			s.Color = "#6366F1"
		}
		err := pool.QueryRow(context.Background(),
			`INSERT INTO stages (event_id, name, color)
			 VALUES ($1,$2,$3)
			 RETURNING id, event_id, name, color, display_order`,
			c.Param("id"), s.Name, s.Color).
			Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, s)
	}
}

func deleteStage(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM stages WHERE id = $1 AND event_id = $2`,
			c.Param("stage_id"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
