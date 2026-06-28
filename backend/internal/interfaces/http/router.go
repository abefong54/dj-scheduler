package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"eventlineup/internal/interfaces/http/middleware"
)

func NewRouter(frontendURL, jwtSecret string, public *PublicHandler, dj *DJHandler, ev *EventHandler, st *StageHandler, sl *SlotHandler) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendURL},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Public routes — no auth required (shareable schedule link).
	publicAPI := r.Group("/api")
	public.Register(publicAPI)

	// Protected routes — every request must carry a valid organizer JWT.
	api := r.Group("/api")
	api.Use(middleware.Auth(jwtSecret))
	dj.Register(api)
	ev.Register(api)
	st.Register(api)
	sl.Register(api)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
