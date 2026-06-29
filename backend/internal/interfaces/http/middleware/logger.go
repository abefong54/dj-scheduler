package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger returns a Gin middleware that logs one line per request using
// the matched route template (c.FullPath(), e.g. "/api/dj/portal") rather than
// the raw request URL.
//
// This deliberately omits query strings: gin.Default()'s logger writes the full
// path including the query, which leaks secret-bearing parameters — the OAuth
// code/state and the DJ portal token — into stdout access logs (EL-037). The
// route template carries no user-supplied values, so it is safe to log.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			// No route matched (e.g. a 404). Log a placeholder rather than the
			// raw URL so an unmatched, secret-bearing path can't leak either.
			route = "unmatched"
		}
		log.Printf("%3d %-7s %s %v %s",
			c.Writer.Status(),
			c.Request.Method,
			route,
			time.Since(start),
			c.ClientIP(),
		)
	}
}
