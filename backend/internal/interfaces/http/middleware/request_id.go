package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestIDKey is the gin context key under which the per-request id is stored.
const RequestIDKey = "request_id"

// requestIDHeader is echoed back so a client (or an upstream proxy) can correlate
// a response with the server-side logs for that request.
const requestIDHeader = "X-Request-ID"

// RequestID returns middleware that assigns each request a random id, stores it
// on the context (for structured logging — see EL-039) and echoes it in the
// X-Request-ID response header. A generated id is used unconditionally rather
// than trusting an inbound header, so a client can't spoof or collide ids.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := newRequestID()
		c.Set(RequestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// GetRequestID returns the request id assigned by RequestID, or "" if none was
// set (e.g. a request that bypassed the middleware).
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// newRequestID returns 16 hex chars of cryptographic randomness. crypto/rand
// never fails on the platforms we target, but if it ever did we fall back to a
// fixed marker rather than panicking on a hot path.
func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}
