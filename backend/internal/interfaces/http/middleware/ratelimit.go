package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor is one client's token-bucket allowance. tokens refill at the limiter's
// rate up to its burst; each allowed request costs one token.
type visitor struct {
	tokens float64
	last   time.Time
}

// rateLimiter is a per-key token-bucket limiter, safe for concurrent use. It is
// deliberately dependency-free (no golang.org/x/time/rate) — the bucket is a few
// lines and keeps the module lean per EL-086.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    float64
}

func newRateLimiter(perSecond float64, burst int) *rateLimiter {
	return &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     perSecond,
		burst:    float64(burst),
	}
}

// allow reports whether a request from key is permitted at time now, consuming a
// token when it is. now is a parameter (not time.Now()) so the behaviour is
// deterministically testable.
func (rl *rateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[key]
	if !ok {
		rl.visitors[key] = &visitor{tokens: rl.burst - 1, last: now}
		return true
	}

	// Refill for the elapsed time, capped at burst.
	v.tokens = min(rl.burst, v.tokens+now.Sub(v.last).Seconds()*rl.rate)
	v.last = now
	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// cleanup evicts visitors idle longer than ttl so the map can't grow unbounded.
func (rl *rateLimiter) cleanup(now time.Time, ttl time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for k, v := range rl.visitors {
		if now.Sub(v.last) > ttl {
			delete(rl.visitors, k)
		}
	}
}

// RateLimit returns middleware that per-client token-bucket rate-limits requests
// and answers a safe 429 (no request detail, nothing logged) once the bucket is
// empty — abuse protection for public write endpoints like POST /api/leads
// (EL-086). perSecond is the sustained refill rate, burst the bucket size.
//
// The key is c.ClientIP(). Behind a proxy whose forwarded headers aren't trusted
// (this app sets SetTrustedProxies(nil)), that resolves to the edge address, so
// the limit is effectively per-edge — still a hard cap on flooding the table.
// Trust the platform's proxy hop to key on the true client instead.
func RateLimit(perSecond float64, burst int) gin.HandlerFunc {
	rl := newRateLimiter(perSecond, burst)

	// Background janitor: drop idle buckets periodically.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for now := range ticker.C {
			rl.cleanup(now, 10*time.Minute)
		}
	}()

	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP(), time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
