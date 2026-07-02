package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_AllowsBurstThenBlocks(t *testing.T) {
	rl := newRateLimiter(1, 3) // 1 token/sec, burst 3
	now := time.Unix(1000, 0)

	for i := 0; i < 3; i++ {
		if !rl.allow("1.2.3.4", now) {
			t.Fatalf("request %d within burst should be allowed", i+1)
		}
	}
	if rl.allow("1.2.3.4", now) {
		t.Fatal("4th request in the same instant should be blocked")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := newRateLimiter(1, 2) // 1 token/sec
	now := time.Unix(2000, 0)

	rl.allow("ip", now)
	rl.allow("ip", now) // bucket now empty
	if rl.allow("ip", now) {
		t.Fatal("should be blocked while empty")
	}
	if !rl.allow("ip", now.Add(time.Second)) {
		t.Fatal("one token should have refilled after 1s")
	}
	if rl.allow("ip", now.Add(time.Second)) {
		t.Fatal("only one token refills in 1s")
	}
}

func TestRateLimiter_PerKeyIndependent(t *testing.T) {
	rl := newRateLimiter(1, 1)
	now := time.Unix(3000, 0)

	if !rl.allow("a", now) || !rl.allow("b", now) {
		t.Fatal("distinct keys must not share a bucket")
	}
	if rl.allow("a", now) {
		t.Fatal("key a is now exhausted")
	}
}

func TestRateLimiter_CleanupEvictsIdle(t *testing.T) {
	rl := newRateLimiter(1, 1)
	now := time.Unix(4000, 0)
	rl.allow("stale", now)
	rl.cleanup(now.Add(20*time.Minute), 10*time.Minute)
	rl.mu.Lock()
	_, present := rl.visitors["stale"]
	rl.mu.Unlock()
	if present {
		t.Fatal("idle visitor should have been evicted")
	}
}

func TestRateLimit_Middleware429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/leads", RateLimit(1, 1), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	call := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/leads", nil)
		req.RemoteAddr = "9.9.9.9:5555"
		r.ServeHTTP(w, req)
		return w
	}

	if got := call().Code; got != http.StatusCreated {
		t.Fatalf("first request code = %d, want 201", got)
	}
	w := call()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("second request code = %d, want 429", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "too many requests" {
		t.Fatalf("429 body = %q, want generic message", w.Body.String())
	}
}
