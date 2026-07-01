package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID_SetsHeaderAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/events", nil)

	var seen string
	handler := RequestID()
	handler(c)
	seen = GetRequestID(c)

	if seen == "" {
		t.Fatal("request id not stored in context")
	}
	if got := w.Header().Get("X-Request-ID"); got != seen {
		t.Fatalf("X-Request-ID header = %q, want %q (matching context)", got, seen)
	}
}

func TestRequestID_IsUniquePerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := RequestID()

	first, _ := gin.CreateTestContext(httptest.NewRecorder())
	first.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	handler(first)

	second, _ := gin.CreateTestContext(httptest.NewRecorder())
	second.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	handler(second)

	if GetRequestID(first) == GetRequestID(second) {
		t.Fatal("request ids collided across requests")
	}
}

func TestGetRequestID_MissingReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	if got := GetRequestID(c); got != "" {
		t.Fatalf("GetRequestID with no id = %q, want empty", got)
	}
}
