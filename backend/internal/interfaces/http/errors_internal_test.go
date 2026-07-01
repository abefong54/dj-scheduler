package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
)

// newTestContext returns a gin.Context backed by a recorder, with a request id
// preset the way the RequestID middleware would.
func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/events/abc", nil)
	c.Set(middleware.RequestIDKey, "req-test-123")
	return c, w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", w.Body.String(), err)
	}
	return body
}

func TestRespondError_Validation(t *testing.T) {
	c, w := newTestContext()
	respondError(c, apperrors.NewValidation("name required"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if got := decodeBody(t, w)["error"]; got != "name required" {
		t.Fatalf("error = %q, want %q", got, "name required")
	}
}

func TestRespondError_Conflict(t *testing.T) {
	c, w := newTestContext()
	respondError(c, &apperrors.ConflictError{Type: "dj_double_booked", ConflictingSlotID: "slot-9"})

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
	body := decodeBody(t, w)
	if body["error"] != "conflict" {
		t.Fatalf("error = %q, want conflict", body["error"])
	}
	if body["type"] != "dj_double_booked" {
		t.Fatalf("type = %q, want dj_double_booked", body["type"])
	}
	if body["conflicting_slot_id"] != "slot-9" {
		t.Fatalf("conflicting_slot_id = %q, want slot-9", body["conflicting_slot_id"])
	}
}

func TestRespondError_NotFound(t *testing.T) {
	c, w := newTestContext()
	respondError(c, apperrors.ErrNotFound)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if got := decodeBody(t, w)["error"]; got != "not found" {
		t.Fatalf("error = %q, want %q", got, "not found")
	}
}

func TestRespondError_WrappedNotFound(t *testing.T) {
	c, w := newTestContext()
	respondError(c, fmt.Errorf("loading event: %w", apperrors.ErrNotFound))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for wrapped ErrNotFound", w.Code)
	}
}

func TestRespondError_Forbidden(t *testing.T) {
	c, w := newTestContext()
	respondError(c, apperrors.ErrForbidden)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	if got := decodeBody(t, w)["error"]; got != "forbidden" {
		t.Fatalf("error = %q, want %q", got, "forbidden")
	}
}

// The core of EL-039: an unknown/internal error must never echo its text to the
// client, and the real error must be logged server-side with the request id.
func TestRespondError_InternalHidesTextAndLogs(t *testing.T) {
	var logBuf bytes.Buffer
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	c, w := newTestContext()
	raw := errors.New("pq: relation \"slots\" does not exist")
	respondError(c, raw)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if got := decodeBody(t, w)["error"]; got != "internal error" {
		t.Fatalf("error = %q, want %q", got, "internal error")
	}
	if strings.Contains(w.Body.String(), "relation") {
		t.Fatalf("response body leaked raw error text: %s", w.Body.String())
	}

	logged := logBuf.String()
	if !strings.Contains(logged, "relation") {
		t.Fatalf("real error was not logged server-side: %s", logged)
	}
	if !strings.Contains(logged, "req-test-123") {
		t.Fatalf("request id not present in server log: %s", logged)
	}
}
