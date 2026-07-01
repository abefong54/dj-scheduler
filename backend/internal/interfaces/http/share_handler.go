package handler

import (
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	eventuc "eventlineup/internal/usecase/event"
	slotuc "eventlineup/internal/usecase/slot"
)

// ShareHandler serves a server-rendered Open Graph / share page for a single
// slot's per-DJ card (EL-049). A pure SPA can't unfurl on LINE or social
// platforms because their crawlers don't run JavaScript, so the share link
// points here: this page carries the og:* meta tags for the crawler and
// immediately redirects a human browser to the SPA card route.
//
// The route is tokenless and public — the slot id in the path is the only
// identifier. All interpolated data (DJ/event/stage names) is user-controlled,
// so the page is rendered with html/template for contextual auto-escaping.
type ShareHandler struct {
	slots       *slotuc.UseCase
	events      *eventuc.UseCase
	frontendURL string
}

func NewShareHandler(slots *slotuc.UseCase, events *eventuc.UseCase, frontendURL string) *ShareHandler {
	return &ShareHandler{slots: slots, events: events, frontendURL: frontendURL}
}

// Register mounts the share route at the gin engine root (not under /api, not
// behind auth), alongside the other root routes like /healthz.
func (h *ShareHandler) Register(r *gin.Engine) {
	r.GET("/s/dj/:slotId", h.card)
}

// shareData is the html/template payload. Every field is auto-escaped in its
// rendering context (HTML attribute or JS string), neutralising XSS from the
// user-controlled names.
type shareData struct {
	Title       string
	Description string
	Image       string
	CardURL     string
}

// shareTmpl renders the OG page. CardURL is interpolated into attribute and JS
// string contexts; html/template escapes each appropriately.
var shareTmpl = template.Must(template.New("share").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:image" content="{{.Image}}">
<meta property="og:url" content="{{.CardURL}}">
<meta property="og:type" content="website">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="{{.Title}}">
<meta name="twitter:description" content="{{.Description}}">
<meta name="twitter:image" content="{{.Image}}">
<link rel="canonical" href="{{.CardURL}}">
<meta http-equiv="refresh" content="0; url={{.CardURL}}">
<script>location.replace("{{.CardURL}}")</script>
</head>
<body>
<a href="{{.CardURL}}">View card</a>
</body>
</html>`))

// notFoundTmpl is a minimal valid HTML body for an unknown slot. Static, so no
// interpolation and no way to crash.
var notFoundTmpl = template.Must(template.New("share404").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Not found</title></head>
<body>Not found</body>
</html>`))

// humanDate formats a YYYY-MM-DD slot date like "Sat 12 Jul". On a parse
// failure it returns the raw input unchanged — never panics.
func humanDate(slotDate string) string {
	t, err := time.Parse("2006-01-02", slotDate)
	if err != nil {
		return slotDate
	}
	return t.Format("Mon 02 Jan")
}

func (h *ShareHandler) card(c *gin.Context) {
	ctx := c.Request.Context()
	slotID := c.Param("slotId")

	slot, err := h.slots.GetPublicByID(ctx, slotID)
	if err != nil {
		h.renderNotFound(c)
		return
	}

	event, err := h.events.GetPublic(ctx, slot.EventID)
	if err != nil {
		h.renderNotFound(c)
		return
	}

	// Build the SPA card URL from a path-escaped slot id so an odd id can't break
	// out of the path; html/template then escapes it per rendering context.
	cardURL := h.frontendURL + "/card/" + url.PathEscape(slotID)

	data := shareData{
		Title:       slot.DjName + " is playing " + event.Name,
		Description: slot.StageName + " · " + slot.StartTime + " · " + humanDate(slot.SlotDate),
		Image:       h.frontendURL + "/assets/og/card-default.png",
		CardURL:     cardURL,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = shareTmpl.Execute(c.Writer, data)
}

func (h *ShareHandler) renderNotFound(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusNotFound)
	_ = notFoundTmpl.Execute(c.Writer, nil)
}
