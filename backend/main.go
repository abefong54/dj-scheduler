package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	pool, err := InitDB(dbURL)
	if err != nil {
		log.Fatalf("InitDB: %v", err)
	}
	defer pool.Close()

	r := gin.Default()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendURL},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	registerRoutes(r, pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}

func registerRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	api := r.Group("/api")
	registerDJRoutes(api, pool)
	registerEventRoutes(api, pool)
	registerStageRoutes(api, pool)
	registerSlotRoutes(api, pool)
}
