package app

import (
	"context"
	"log"
	"time"

	"eventlineup/internal/infrastructure/config"
	"eventlineup/internal/infrastructure/database"
	"eventlineup/internal/infrastructure/googleauth"
	"eventlineup/internal/infrastructure/lineauth"
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

// googleStudentAdapter adapts the shared OAuth provider (which returns
// authuc.ExternalUser) to the studentauth.GoogleAuthenticator interface, so
// student sign-in reuses the same Google client as the organizer flow. The
// provider's Subject is the stable Google user id. The student callback is a
// distinct redirect URI — Track A confirms it's registered.
type googleStudentAdapter struct{ inner authuc.OAuthProvider }

func (a googleStudentAdapter) AuthCodeURL(state string) string { return a.inner.AuthCodeURL(state) }

func (a googleStudentAdapter) UserFromCode(ctx context.Context, code string) (studentauthuc.GoogleUser, error) {
	u, err := a.inner.UserFromCode(ctx, code)
	if err != nil {
		return studentauthuc.GoogleUser{}, err
	}
	return studentauthuc.GoogleUser{GoogleID: u.Subject, Email: u.Email, Name: u.Name}, nil
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
	if err := cfg.ValidateLine(); err != nil {
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
	authProviders := map[string]authuc.OAuthProvider{authuc.ProviderGoogle: googleAuth}
	authProviderNames := []string{authuc.ProviderGoogle}
	// LINE Login (EL-110) is opt-in: only wired when fully configured, so local
	// dev without LINE creds still boots. Google stays the required provider.
	if cfg.LineLoginConfigured() {
		authProviders[authuc.ProviderLine] = lineauth.New(cfg.LineChannelID, cfg.LineChannelSecret, cfg.LineRedirectURL)
		authProviderNames = append(authProviderNames, authuc.ProviderLine)
	}
	authHandler := httphandler.NewAuthHandler(authuc.New(authProviders, organizerRepo, membershipRepo, cfg.JWTSecret, tokenTTL), cfg.FrontendURL, cfg.SecureCookies, authProviderNames)
	// Onboarding surface: /api/me, register school, invite members (EL-105).
	schoolHandler := httphandler.NewSchoolHandler(schooluc.New(organizerRepo, schoolRepo, membershipRepo))

	// Student features (2026-07-05 sprint). Student auth reuses the Google client
	// (via adapter) + djRepo; the challenge and submission usecases get their real
	// school-scoped repos (Tracks B/C).
	studentAuthHandler := httphandler.NewStudentAuthHandler(
		studentauthuc.New(googleStudentAdapter{inner: googleAuth}, djRepo, cfg.JWTSecret, tokenTTL),
		cfg.FrontendURL, cfg.SecureCookies,
	)
	challengeHandler := httphandler.NewChallengeHandler(challengeuc.New(database.NewChallengeRepository(pool)))
	submissionHandler := httphandler.NewSubmissionHandler(submissionuc.New(database.NewSubmissionRepository(pool)))

	// membershipRepo also resolves the caller's school for the scoping middleware (EL-104).
	r := httphandler.NewRouter(cfg.FrontendURL, cfg.JWTSecret, membershipRepo, publicHandler, shareHandler, djPortalHandler, djHandler, eventHandler, stageHandler, slotHandler, lineHandler, perfHandler, leadHandler, schoolHandler, studentAuthHandler, challengeHandler, submissionHandler)
	authHandler.Register(r) // unauthenticated auth routes

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
