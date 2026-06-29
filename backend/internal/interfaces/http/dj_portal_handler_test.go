package handler_test

import (
	"bytes"
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
	slotuc "eventlineup/internal/usecase/slot"
)

// djPortalRouters builds two engines sharing one DJ usecase: an organizer-scoped
// one for minting tokens, and a public one (no auth) for redeeming them.
func djPortalRouters(t *testing.T, pool *pgxpool.Pool, organizerID string) (org *gin.Engine, public *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := httphandler.NewDJPortalHandler(
		djuc.New(database.NewDJRepository(pool)),
		slotuc.New(database.NewSlotRepository(pool)),
		"http://localhost:4200",
	)

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

// patchConfirm PATCHes a slot's confirmation via the public portal route.
func patchConfirm(public *gin.Engine, slotID, token, confirmation string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"confirmation":"` + confirmation + `"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/dj/portal/slots/"+slotID+"?token="+token, body)
	req.Header.Set("Content-Type", "application/json")
	public.ServeHTTP(w, req)
	return w
}

// seedSlotForDJ inserts a slot booked for djID and returns its ID.
func seedSlotForDJ(t *testing.T, pool *pgxpool.Pool, eventID, stageID, djID string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO slots (event_id, stage_id, dj_id, slot_date, start_time, end_time)
		 VALUES ($1, $2, $3, '2026-07-25', '20:00', '21:00') RETURNING id`,
		eventID, stageID, djID).Scan(&id); err != nil {
		t.Fatalf("seed slot: %v", err)
	}
	return id
}

// US-011: DJ confirms then re-flags a slot; status persists and is reflected on the portal.
func TestDJPortalConfirmSlot(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, org, "Sasha")
	eventID := createTestEventForOrg(t, pool, org)
	stageID := createTestStage(t, pool, eventID)
	slotID := seedSlotForDJ(t, pool, eventID, stageID, djID)

	orgR, publicR := djPortalRouters(t, pool, org)
	token, _ := mintPortalToken(t, orgR, djID)

	dbConfirmation := func() *string {
		var v *string
		pool.QueryRow(context.Background(), "SELECT dj_confirmation FROM slots WHERE id = $1", slotID).Scan(&v)
		return v
	}

	// Confirm.
	if w := patchConfirm(publicR, slotID, token, "confirmed"); w.Code != http.StatusOK {
		t.Fatalf("confirm: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if v := dbConfirmation(); v == nil || *v != "confirmed" {
		t.Fatalf("expected DB confirmation 'confirmed', got %v", v)
	}

	// Change the response to flagged.
	if w := patchConfirm(publicR, slotID, token, "flagged"); w.Code != http.StatusOK {
		t.Fatalf("flag: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if v := dbConfirmation(); v == nil || *v != "flagged" {
		t.Fatalf("expected DB confirmation 'flagged', got %v", v)
	}

	// Portal GET reflects the confirmation.
	w := getPortal(publicR, token)
	var resp struct {
		Slots []struct {
			DJConfirmation *string `json:"dj_confirmation"`
		} `json:"slots"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Slots) != 1 || resp.Slots[0].DJConfirmation == nil || *resp.Slots[0].DJConfirmation != "flagged" {
		t.Fatalf("portal should show flagged confirmation, got %+v", resp.Slots)
	}
}

func TestDJPortalConfirmInvalidValue(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djID := seedDJ(t, pool, org, "Sasha")
	eventID := createTestEventForOrg(t, pool, org)
	stageID := createTestStage(t, pool, eventID)
	slotID := seedSlotForDJ(t, pool, eventID, stageID, djID)

	orgR, publicR := djPortalRouters(t, pool, org)
	token, _ := mintPortalToken(t, orgR, djID)

	if w := patchConfirm(publicR, slotID, token, "maybe"); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid value: expected 400, got %d", w.Code)
	}
	if w := patchConfirm(publicR, slotID, "", "confirmed"); w.Code != http.StatusUnauthorized {
		t.Fatalf("empty token: expected 401, got %d", w.Code)
	}
	if w := patchConfirm(publicR, slotID, token+"bad", "confirmed"); w.Code != http.StatusUnauthorized {
		t.Fatalf("bad token: expected 401, got %d", w.Code)
	}
}

// US-011: a DJ's token cannot confirm a slot that belongs to a different DJ.
func TestDJPortalConfirmWrongDJ(t *testing.T) {
	pool := setupTestDB(t)
	org := createTestOrganizer(t, pool)
	djA := seedDJ(t, pool, org, "Sasha")
	djB := seedDJ(t, pool, org, "Marco")
	eventID := createTestEventForOrg(t, pool, org)
	stageID := createTestStage(t, pool, eventID)
	slotA := seedSlotForDJ(t, pool, eventID, stageID, djA) // belongs to djA

	orgR, publicR := djPortalRouters(t, pool, org)
	tokenB, _ := mintPortalToken(t, orgR, djB) // djB's token

	if w := patchConfirm(publicR, slotA, tokenB, "confirmed"); w.Code != http.StatusForbidden {
		t.Fatalf("wrong-DJ confirm: expected 403, got %d: %s", w.Code, w.Body.String())
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
