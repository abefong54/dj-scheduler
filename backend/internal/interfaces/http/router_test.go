package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	"eventlineup/internal/interfaces/http/middleware"
	djuc "eventlineup/internal/usecase/dj"
	eventuc "eventlineup/internal/usecase/event"
	slotuc "eventlineup/internal/usecase/slot"
	stageuc "eventlineup/internal/usecase/stage"
)

const routerTestSecret = "router-test-secret-router-test-secret-0123"

func fullRouter(t *testing.T) *gin.Engine {
	t.Helper()
	pool := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	dj := httphandler.NewDJHandler(djuc.New(database.NewDJRepository(pool)))
	ev := httphandler.NewEventHandler(eventuc.New(database.NewEventRepository(pool)))
	st := httphandler.NewStageHandler(stageuc.New(database.NewStageRepository(pool)))
	sl := httphandler.NewSlotHandler(slotuc.New(database.NewSlotRepository(pool)))
	pub := httphandler.NewPublicHandler(
		eventuc.New(database.NewEventRepository(pool)),
		stageuc.New(database.NewStageRepository(pool)),
		slotuc.New(database.NewSlotRepository(pool)),
	)
	return httphandler.NewRouter("http://localhost:4200", routerTestSecret, pub, dj, ev, st, sl)
}

func mintRouterToken(t *testing.T) string {
	t.Helper()
	claims := middleware.Claims{
		OrganizerID: "org-router",
		Email:       "organizer@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(routerTestSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func TestRouterProtectedRouteRejectsWithoutToken(t *testing.T) {
	r := fullRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouterProtectedRouteAcceptsValidToken(t *testing.T) {
	r := fullRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Header.Set("Authorization", "Bearer "+mintRouterToken(t))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouterPublicRouteBypassesAuth(t *testing.T) {
	pool := setupTestDB(t)
	r := fullRouter(t)
	eventID := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/public", nil)
	// No Authorization header.
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on public route without token, got %d: %s", w.Code, w.Body.String())
	}
}
