package app

import (
	"context"
	"log"
	"time"

	"eventlineup/internal/infrastructure/config"
	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/infrastructure/googleauth"
	httphandler "eventlineup/internal/interfaces/http"
	authuc "eventlineup/internal/usecase/auth"
	challengeuc "eventlineup/internal/usecase/challenge"
	djuc "eventlineup/internal/usecase/dj"
	eventuc "eventlineup/internal/usecase/event"
	leaduc "eventlineup/internal/usecase/lead"
	linenotifyuc "eventlineup/internal/usecase/linenotify"
	perfuc "eventlineup/internal/usecase/performance"
	schooluc "eventlineup/internal/usecase/school"
	slotuc "eventlineup/internal/usecase/slot"
	stageuc "eventlineup/internal/usecase/stage"
	studentauthuc "eventlineup/internal/usecase/studentauth"
	submissionuc "eventlineup/internal/usecase/submission"
)

// googleStudentAdapter adapts the shared Google client (which returns
// authuc.GoogleUser) to the studentauth.GoogleAuthenticator interface, so student
// sign-in reuses the same OAuth client as the organizer flow. The student
// callback is a distinct redirect URI — Track A confirms it's registered.
type googleStudentAdapter struct{ inner authuc.GoogleAuthenticator }

func (a googleStudentAdapter) AuthCodeURL(state string) string { return a.inner.AuthCodeURL(state) }

func (a googleStudentAdapter) UserFromCode(ctx context.Context, code string) (studentauthuc.GoogleUser, error) {
	u, err := a.inner.UserFromCode(ctx, code)
	if err != nil {
		return studentauthuc.GoogleUser{}, err
	}
	return studentauthuc.GoogleUser{GoogleID: u.GoogleID, Email: u.Email, Name: u.Name}, nil
}

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
	if err := cfg.ValidateLineNotify(); err != nil {
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
	perfRepo := database.NewPerformanceRepository(pool)

	djUC := djuc.New(djRepo)
	slotUC := slotuc.New(slotRepo)
	djHandler := httphandler.NewDJHandler(djUC)
	djPortalHandler := httphandler.NewDJPortalHandler(djUC, slotUC, cfg.FrontendURL)
	eventHandler := httphandler.NewEventHandler(eventuc.New(eventRepo))
	stageHandler := httphandler.NewStageHandler(stageuc.New(stageRepo))
	slotHandler := httphandler.NewSlotHandler(slotUC)
	lineHandler := httphandler.NewLineHandler(linenotifyuc.New(eventRepo, cfg.LineNotifyEncryptionKey))
	perfHandler := httphandler.NewPerformanceHandler(perfuc.New(perfRepo))
	leadRepo := database.NewLeadRepository(pool)
	leadHandler := httphandler.NewLeadHandler(leaduc.New(leadRepo))
	publicHandler := httphandler.NewPublicHandler(eventuc.New(eventRepo), stageuc.New(stageRepo), slotuc.New(slotRepo))
	shareHandler := httphandler.NewShareHandler(slotuc.New(slotRepo), eventuc.New(eventRepo), cfg.FrontendURL)

	organizerRepo := database.NewOrganizerRepository(pool)
	membershipRepo := database.NewMembershipRepository(pool)
	schoolRepo := database.NewSchoolRepository(pool)
	googleAuth := googleauth.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	authHandler := httphandler.NewAuthHandler(authuc.New(googleAuth, organizerRepo, membershipRepo, cfg.JWTSecret, tokenTTL), cfg.FrontendURL, cfg.SecureCookies)
	// Onboarding surface: /api/me, register school, invite members (EL-105).
	schoolHandler := httphandler.NewSchoolHandler(schooluc.New(organizerRepo, schoolRepo, membershipRepo))

	// Student features (2026-07-05 sprint). Student auth reuses the Google client
	// (via adapter) + djRepo; challenge/submission repos are Wave-1 deliverables
	// (Tracks B/C), so their usecases wire nil for now — every route returns 501
	// until filled, so the nil repo is never dereferenced.
	studentAuthHandler := httphandler.NewStudentAuthHandler(
		studentauthuc.New(googleStudentAdapter{inner: googleAuth}, djRepo, cfg.JWTSecret, tokenTTL),
		cfg.FrontendURL, cfg.SecureCookies,
	)
	challengeHandler := httphandler.NewChallengeHandler(challengeuc.New(database.NewChallengeRepository(pool)))
	submissionHandler := httphandler.NewSubmissionHandler(submissionuc.New(nil))

	// membershipRepo also resolves the caller's school for the scoping middleware (EL-104).
	r := httphandler.NewRouter(cfg.FrontendURL, cfg.JWTSecret, membershipRepo, publicHandler, shareHandler, djPortalHandler, djHandler, eventHandler, stageHandler, slotHandler, lineHandler, perfHandler, leadHandler, schoolHandler, studentAuthHandler, challengeHandler, submissionHandler)
	authHandler.Register(r) // unauthenticated auth routes

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
