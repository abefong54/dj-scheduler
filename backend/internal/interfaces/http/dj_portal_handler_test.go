package handler_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/infrastructure/database"
	httphandler "eventlineup/internal/interfaces/http"
	djuc "eventlineup/internal/usecase/dj"
)

// djPortalRouters builds two engines sharing one DJ usecase: an organizer-scoped
// one for minting tokens, and a public one (no auth) for redeeming them.
func djPortalRouters(t *testing.T, pool *pgxpool.Pool, organizerID string) (org *gin.Engine, public *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := httphandler.NewDJPortalHandler(djuc.New(database.NewDJRepository(pool)), "http://localhost:4200")

	org = gin.New()
	org.Use(orgContext(organizerID))
	h.RegisterProtected(org.Group("/api"))

	public = gin.New()
	h.RegisterPublic(public.Group("/api"))
	return org, public
}

// mintPortalToken POSTs to the token route and returns the raw token + status.
func mintPortalToken(t *testing.T, org *gin.Engine, djID string) (string, int) {
	t.Helper()
	w := httptest.NewRecorder()
	org.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/djs/"+djID+"/token", nil))
	if w.Code != http.StatusOK {
		return "", w.Code
	}
	var resp struct {
		PortalURL string `json:"portal_url"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	_, token, _ := strings.Cut(resp.PortalURL, "token=")
	return token, w.Code
}

func getPortal(public *gin.Engine, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	public.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/dj/portal?token="+token, nil))
	return w
}

func TestDJPortalTokenAndAccess(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, org, "Sasha")

	// Give the DJ one booked slot in an event the organizer owns.
	eventID := createTestEventForOrg(t, pool, org)
	stageID := createTestStage(t, pool, eventID)
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO slots (event_id, stage_id, dj_id, slot_date, start_time, end_time)
		 VALUES ($1, $2, $3, '2026-07-25', '20:00', '21:00')`,
		eventID, stageID, djID); err != nil {
		t.Fatalf("seed slot: %v", err)
	}

	orgR, publicR := djPortalRouters(t, pool, org)

	token, code := mintPortalToken(t, orgR, djID)
	if code != http.StatusOK {
		t.Fatalf("token generation: expected 200, got %d", code)
	}
	if token == "" {
		t.Fatal("expected a non-empty portal token")
	}

	// Valid token returns the DJ and their slots.
	w := getPortal(publicR, token)
	if w.Code != http.StatusOK {
		t.Fatalf("portal with valid token: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		DJ struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"dj"`
		Slots []struct {
			EventName string `json:"event_name"`
			StageName string `json:"stage_name"`
		} `json:"slots"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode portal: %v", err)
	}
	if resp.DJ.ID != djID || resp.DJ.Name != "Sasha" {
		t.Fatalf("expected DJ %s/Sasha, got %s/%s", djID, resp.DJ.ID, resp.DJ.Name)
	}
	if len(resp.Slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(resp.Slots))
	}
	if resp.Slots[0].EventName == "" || resp.Slots[0].StageName == "" {
		t.Fatalf("expected event_name and stage_name on slot, got %+v", resp.Slots[0])
	}

	// Tampered token → 401.
	if w := getPortal(publicR, token+"deadbeef"); w.Code != http.StatusUnauthorized {
		t.Fatalf("tampered token: expected 401, got %d", w.Code)
	}
	// Empty token → 401.
	if w := getPortal(publicR, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("empty token: expected 401, got %d", w.Code)
	}

	// Regenerating revokes the old token.
	newToken, code := mintPortalToken(t, orgR, djID)
	if code != http.StatusOK {
		t.Fatalf("regenerate: expected 200, got %d", code)
	}
	if newToken == token {
		t.Fatal("regenerated token should differ from the original")
	}
	if w := getPortal(publicR, token); w.Code != http.StatusUnauthorized {
		t.Fatalf("old token after regenerate: expected 401, got %d", w.Code)
	}
	if w := getPortal(publicR, newToken); w.Code != http.StatusOK {
		t.Fatalf("new token: expected 200, got %d", w.Code)
	}
}

func TestDJPortalExpiredToken(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, org, "Marco")

	// Store a token hash that expired yesterday.
	raw := "expired-raw-token-value"
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])
	if _, err := pool.Exec(context.Background(),
		`UPDATE djs SET portal_token_hash = $1, portal_token_expires_at = $2 WHERE id = $3`,
		hash, time.Now().Add(-24*time.Hour), djID); err != nil {
		t.Fatalf("set expired token: %v", err)
	}

	_, publicR := djPortalRouters(t, pool, org)
	if w := getPortal(publicR, raw); w.Code != http.StatusUnauthorized {
		t.Fatalf("expired token: expected 401, got %d", w.Code)
	}
}

// US-009 / US-002: minting a token for another organizer's DJ returns 404.
func TestDJPortalTokenCrossOrganizer(t *testing.T) {
	pool := setupTestDB(t)
	orgA := createTestOrganizer(t, pool)
	orgB := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, orgA, "Kay")

	orgBRouter, _ := djPortalRouters(t, pool, orgB)
	w := httptest.NewRecorder()
	orgBRouter.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/djs/"+djID+"/token", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("cross-organizer token: expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
