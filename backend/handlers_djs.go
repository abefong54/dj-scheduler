// backend/handlers_djs.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerDJRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/djs", listDJs(pool))
	rg.POST("/djs", createDJ(pool))
	rg.DELETE("/djs/:id", deleteDJ(pool))
}

func listDJs(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text FROM djs ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		djs := []DJ{}
		for rows.Next() {
			var d DJ
			if err := rows.Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			djs = append(djs, d)
		}
		c.JSON(http.StatusOK, djs)
	}
}

func createDJ(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var d DJ
		if err := c.ShouldBindJSON(&d); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if d.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
			return
		}
		if d.GenreTags == nil {
			d.GenreTags = []string{}
		}
		err := pool.QueryRow(context.Background(),
			`INSERT INTO djs (name, genre_tags) VALUES ($1, $2)
			 RETURNING id, name, COALESCE(genre_tags, '{}'), created_at::text`,
			d.Name, d.GenreTags).Scan(&d.ID, &d.Name, &d.GenreTags, &d.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, d)
	}
}

func deleteDJ(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM djs WHERE id = $1`, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
