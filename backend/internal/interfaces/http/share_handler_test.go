package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	httphandler "eventlineup/internal/interfaces/http"
	eventuc "eventlineup/internal/usecase/event"
	slotuc "eventlineup/internal/usecase/slot"
)

// shareFakeSlotRepo is an in-memory SlotRepository for share-handler tests, so
// the XSS guard runs without a database. Only GetPublicByID is exercised.
type shareFakeSlotRepo struct {
	slot model.Slot
	err  error
}

func (r *shareFakeSlotRepo) List(context.Context, string, string) ([]model.Slot, error) {
	return nil, nil
}
func (r *shareFakeSlotRepo) ListPublic(context.Context, string) ([]model.Slot, error) {
	return nil, nil
}
func (r *shareFakeSlotRepo) GetPublicByID(context.Context, string) (model.Slot, error) {
	return r.slot, r.err
}
func (r *shareFakeSlotRepo) Get(context.Context, string, string, string) (model.Slot, error) {
	return model.Slot{}, nil
}
func (r *shareFakeSlotRepo) Create(_ context.Context, s model.Slot, _, _ string) (model.Slot, error) {
	return s, nil
}
func (r *shareFakeSlotRepo) Update(_ context.Context, s model.Slot, _, _ string) (model.Slot, error) {
	return s, nil
}
func (r *shareFakeSlotRepo) Delete(context.Context, string, string, string) error { return nil }
func (r *shareFakeSlotRepo) SetDJConfirmation(context.Context, string, string, string) error {
	return nil
}

// shareFakeEventRepo is an in-memory EventRepository for share-handler tests.
type shareFakeEventRepo struct {
	event model.Event
	err   error
}

func (r *shareFakeEventRepo) List(context.Context, string) ([]model.Event, error) { return nil, nil }
func (r *shareFakeEventRepo) Get(context.Context, string, string) (model.Event, error) {
	return model.Event{}, nil
}
func (r *shareFakeEventRepo) GetPublic(context.Context, string) (model.Event, error) {
	return r.event, r.err
}
func (r *shareFakeEventRepo) Create(_ context.Context, e model.Event, _ string) (model.Event, error) {
	return e, nil
}
func (r *shareFakeEventRepo) Update(_ context.Context, e model.Event, _ string) (model.Event, error) {
	return e, nil
}
func (r *shareFakeEventRepo) Delete(context.Context, string, string) error { return nil }
func (r *shareFakeEventRepo) Clone(context.Context, string, string) (model.Event, error) {
	return model.Event{}, nil
}
func (r *shareFakeEventRepo) SetLineNotifyToken(context.Context, string, string, *string) (bool, error) {
	return false, nil
}

func shareRouter(t *testing.T, slotRepo *shareFakeSlotRepo, eventRepo *shareFakeEventRepo) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httphandler.NewShareHandler(
		slotuc.New(slotRepo),
		eventuc.New(eventRepo),
		"https://app.example.com",
	)
	h.Register(r)
	return r
}

func TestShareCardRendersOGTags(t *testing.T) {
	slotRepo := &shareFakeSlotRepo{slot: model.Slot{
		ID: "slot-42", EventID: "evt-1", StageName: "Main Stage",
		DjName: "DJ Sasha", SlotDate: "2026-07-12", StartTime: "22:00", EndTime: "23:00",
	}}
	eventRepo := &shareFakeEventRepo{event: model.Event{
		ID: "evt-1", Name: "Summer Fest", VenueName: "The Warehouse",
	}}
	r := shareRouter(t, slotRepo, eventRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/s/dj/slot-42", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("expected text/html content type, got %q", ct)
	}
	body := w.Body.String()

	wantSubstrings := []string{
		`<meta property="og:title" content="DJ Sasha is playing Summer Fest">`,
		`<meta property="og:description" content="Main Stage · 22:00 · Sun 12 Jul">`,
		`<meta property="og:image" content="https://app.example.com/assets/og/card-default.png">`,
		`<meta property="og:url" content="https://app.example.com/card/slot-42">`,
		`<meta property="og:type" content="website">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<link rel="canonical" href="https://app.example.com/card/slot-42">`,
		`<meta http-equiv="refresh" content="0; url=https://app.example.com/card/slot-42">`,
		// html/template escapes the URL for JS-string context inside <script>, so
		// the slashes are backslash-escaped. This is the safe, expected form.
		`location.replace("https:\/\/app.example.com\/card\/slot-42")`,
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q\n--- body ---\n%s", want, body)
		}
	}
}

// TestShareCardEscapesUserContent is the XSS guard: a DJ name containing markup
// must never appear raw in the response — html/template must escape it.
func TestShareCardEscapesUserContent(t *testing.T) {
	slotRepo := &shareFakeSlotRepo{slot: model.Slot{
		ID: "slot-1", EventID: "evt-1", StageName: "Main",
		DjName: `<script>alert(1)</script>`, SlotDate: "2026-07-12", StartTime: "22:00",
	}}
	eventRepo := &shareFakeEventRepo{event: model.Event{
		ID: "evt-1", Name: "Tom & Jerry's Bash",
	}}
	r := shareRouter(t, slotRepo, eventRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/s/dj/slot-1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()

	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Errorf("unescaped <script> from DJ name leaked into page:\n%s", body)
	}
	if strings.Contains(body, "Tom & Jerry's Bash") {
		t.Errorf("raw ampersand from event name not escaped:\n%s", body)
	}
	// The escaped form must still be present, proving the data rendered.
	if !strings.Contains(body, "Tom &amp; Jerry") {
		t.Errorf("expected HTML-escaped event name, not found:\n%s", body)
	}
}

func TestShareCardSlotNotFoundReturns404(t *testing.T) {
	slotRepo := &shareFakeSlotRepo{err: apperrors.ErrNotFound}
	eventRepo := &shareFakeEventRepo{}
	r := shareRouter(t, slotRepo, eventRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/s/dj/missing", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Not found") {
		t.Errorf("expected minimal not-found HTML body, got:\n%s", w.Body.String())
	}
}
