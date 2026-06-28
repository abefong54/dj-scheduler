package app

import (
	"log"

	"eventlineup/internal/infrastructure/config"
	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	djuc "eventlineup/internal/usecase/dj"
	eventuc "eventlineup/internal/usecase/event"
	stageuc "eventlineup/internal/usecase/stage"
	slotuc "eventlineup/internal/usecase/slot"
)

func Run() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	if err := cfg.Validate(); err != nil {
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

	djHandler := httphandler.NewDJHandler(djuc.New(djRepo))
	eventHandler := httphandler.NewEventHandler(eventuc.New(eventRepo))
	stageHandler := httphandler.NewStageHandler(stageuc.New(stageRepo))
	slotHandler := httphandler.NewSlotHandler(slotuc.New(slotRepo))
	publicHandler := httphandler.NewPublicHandler(eventuc.New(eventRepo), stageuc.New(stageRepo), slotuc.New(slotRepo))

	r := httphandler.NewRouter(cfg.FrontendURL, cfg.JWTSecret, publicHandler, djHandler, eventHandler, stageHandler, slotHandler)

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
