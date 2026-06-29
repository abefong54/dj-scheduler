package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"eventlineup/internal/interfaces/http/middleware"
)

func NewRouter(frontendURL, jwtSecret string, public *PublicHandler, djPortal *DJPortalHandler, dj *DJHandler, ev *EventHandler, st *StageHandler, sl *SlotHandler) *gin.Engine {
	// gin.New() (not gin.Default()) so we control logging: the default logger
	// writes the full request URL including query strings, leaking OAuth
	// code/state and DJ portal tokens into access logs (EL-037). RequestLogger
	// logs the route template instead.
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendURL},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Public routes — no auth required (shareable schedule link, DJ portal token).
	publicAPI := r.Group("/api")
	public.Register(publicAPI)
	djPortal.RegisterPublic(publicAPI)

	// Protected routes — every request must carry a valid organizer JWT.
	api := r.Group("/api")
	api.Use(middleware.Auth(jwtSecret))
	djPortal.RegisterProtected(api)
	dj.Register(api)
	ev.Register(api)
	st.Register(api)
	sl.Register(api)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
