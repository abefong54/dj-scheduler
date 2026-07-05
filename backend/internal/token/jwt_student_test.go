package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignStudentSetsRoleAndScope(t *testing.T) {
	signed, err := SignStudent(tokenTestSecret, "dj-7", "school-9", time.Hour)
	if err != nil {
		t.Fatalf("SignStudent: %v", err)
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
	if claims.Role != RoleStudent {
		t.Fatalf("role = %q, want %q", claims.Role, RoleStudent)
	}
	if claims.SchoolID != "school-9" {
		t.Fatalf("school_id = %q, want school-9", claims.SchoolID)
	}
	if claims.Subject != "dj-7" {
		t.Fatalf("subject = %q, want dj-7 (the DJ id)", claims.Subject)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected exp claim to be set")
	}
}

func TestSignStudentDistinctFromOrganizer(t *testing.T) {
	// The organizer Sign path must not mint a student role.
	signed, err := Sign(tokenTestSecret, "org-1", "e@x.com", time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	var claims Claims
	if _, err := jwt.ParseWithClaims(signed, &claims, func(*jwt.Token) (any, error) {
		return []byte(tokenTestSecret), nil
	}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Role == RoleStudent {
		t.Fatal("organizer token must not carry the student role")
	}
}
