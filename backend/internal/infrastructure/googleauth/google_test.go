package googleauth

import (
	"strings"
	"testing"
)

func TestAuthCodeURLPointsAtGoogle(t *testing.T) {
	a := New("test-client-id", "test-secret", "http://localhost:8080/auth/google/callback")
	url := a.AuthCodeURL("xyz-state")

	if !strings.HasPrefix(url, "https://accounts.google.com/o/oauth2/auth") {
		t.Fatalf("expected Google consent URL, got %s", url)
	}
	for _, want := range []string{"state=xyz-state", "client_id=test-client-id", "scope=", "redirect_uri="} {
		if !strings.Contains(url, want) {
			t.Fatalf("expected URL to contain %q, got %s", want, url)
		}
	}
}
