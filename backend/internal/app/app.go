package app

import (
	"log"
	"time"

	"eventlineup/internal/infrastructure/config"
	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/infrastructure/googleauth"
	httphandler "eventlineup/internal/interfaces/http"
	authuc "eventlineup/internal/usecase/auth"
	djuc "eventlineup/internal/usecase/dj"
	eventuc "eventlineup/internal/usecase/event"
	stageuc "eventlineup/internal/usecase/stage"
	slotuc "eventlineup/internal/usecase/slot"
)

// tokenTTL is how long an issued organizer JWT stays valid.
const tokenTTL = 24 * time.Hour

func Run() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	if err := cfg.ValidateGoogle(); err != nil {
		log.Fatal(err)
	}

	pool, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("InitDB: %v", err)
	}
	defer pool.Close()

	djRepo := database.NewDJRepository(pool)
	eventRepo := database.NewEventRepository(pool)
	stageRepo := database.NewStageRepository(pool)
	slotRepo := database.NewSlotRepository(pool)

	djUC := djuc.New(djRepo)
	slotUC := slotuc.New(slotRepo)
	djHandler := httphandler.NewDJHandler(djUC)
	djPortalHandler := httphandler.NewDJPortalHandler(djUC, slotUC, cfg.FrontendURL)
	eventHandler := httphandler.NewEventHandler(eventuc.New(eventRepo))
	stageHandler := httphandler.NewStageHandler(stageuc.New(stageRepo))
	slotHandler := httphandler.NewSlotHandler(slotUC)
	publicHandler := httphandler.NewPublicHandler(eventuc.New(eventRepo), stageuc.New(stageRepo), slotuc.New(slotRepo))

	organizerRepo := database.NewOrganizerRepository(pool)
	googleAuth := googleauth.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	authHandler := httphandler.NewAuthHandler(authuc.New(googleAuth, organizerRepo, cfg.JWTSecret, tokenTTL), cfg.FrontendURL, cfg.SecureCookies)

	r := httphandler.NewRouter(cfg.FrontendURL, cfg.JWTSecret, publicHandler, djPortalHandler, djHandler, eventHandler, stageHandler, slotHandler)
	authHandler.Register(r) // unauthenticated auth routes

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
