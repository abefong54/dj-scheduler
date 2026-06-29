package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// apiContentSecurityPolicy locks the API down to nothing executable. The API
// only ever returns JSON, so it never needs to load scripts, styles, frames, or
// any other resource. If a browser is ever tricked into rendering an API
// response as a document, this CSP ensures nothing runs. The app's own
// (script-bearing) Content-Security-Policy is served at the host/CDN layer in
// front of the Angular bundle (Netlify _headers), not here.
const apiContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'"

// SecurityHeaders returns a Gin middleware that attaches defense-in-depth
// security response headers to every API response (EL-040). It is additive: it
// sets headers and calls the next handler without touching the body or status.
//
//   - X-Content-Type-Options: nosniff — stop MIME-type sniffing of responses.
//   - X-Frame-Options: DENY + CSP frame-ancestors 'none' — the API must never be
//     framed (clickjacking defense; the two cover old and modern browsers).
//   - Referrer-Policy: no-referrer — never leak the request URL in a Referer
//     header, matching the frontend's no-referrer posture (EL-038).
//
// The Swagger UI (/swagger/*) is exempt from the strict CSP because it is an
// interactive HTML app that loads its own scripts and styles; the always-safe
// headers still apply to it.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		if !strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			h.Set("Content-Security-Policy", apiContentSecurityPolicy)
		}
		c.Next()
	}
}
