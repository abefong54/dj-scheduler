package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTestSecret = "token-test-secret-token-test-secret-0123456789"

func TestSignProducesVerifiableToken(t *testing.T) {
	signed, err := Sign(tokenTestSecret, "org-42", "user@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	var claims Claims
	parsed, err := jwt.ParseWithClaims(signed, &claims, func(*jwt.Token) (any, error) {
		return []byte(tokenTestSecret), nil
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected valid token")
	}
	if claims.OrganizerID != "org-42" {
		t.Fatalf("expected organizer_id 'org-42', got %q", claims.OrganizerID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email echoed, got %q", claims.Email)
	}
}

func TestSignSetsExpiry(t *testing.T) {
	signed, err := Sign(tokenTestSecret, "org-1", "user@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	var claims Claims
	if _, err := jwt.ParseWithClaims(signed, &claims, func(*jwt.Token) (any, error) {
		return []byte(tokenTestSecret), nil
	}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected exp claim to be set")
	}
}

func TestSignRejectedByWrongSecret(t *testing.T) {
	signed, err := Sign(tokenTestSecret, "org-1", "user@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	var claims Claims
	_, err = jwt.ParseWithClaims(signed, &claims, func(*jwt.Token) (any, error) {
		return []byte("a-different-secret-a-different-secret-0123"), nil
	})
	if err == nil {
		t.Fatal("expected verification to fail with wrong secret")
	}
}
