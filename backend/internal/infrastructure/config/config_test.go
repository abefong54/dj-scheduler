package config

import (
	"strings"
	"testing"
)

func TestValidateGoogle(t *testing.T) {
	full := Config{
		GoogleClientID:     "id",
		GoogleClientSecret: "secret",
		GoogleRedirectURL:  "http://localhost:8080/auth/google/callback",
	}
	if err := full.ValidateGoogle(); err != nil {
		t.Fatalf("expected complete Google config to validate, got %v", err)
	}
	for _, missing := range []Config{
		{GoogleClientSecret: "secret", GoogleRedirectURL: "x"},
		{GoogleClientID: "id", GoogleRedirectURL: "x"},
		{GoogleClientID: "id", GoogleClientSecret: "secret"},
	} {
		if err := missing.ValidateGoogle(); err == nil {
			t.Fatalf("expected error for incomplete Google config: %+v", missing)
		}
	}
}

func TestLoadReadsJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "a-sufficiently-long-jwt-secret-0123456789")
	cfg := Load()
	if cfg.JWTSecret != "a-sufficiently-long-jwt-secret-0123456789" {
		t.Fatalf("expected JWTSecret loaded from env, got %q", cfg.JWTSecret)
	}
}

func TestLoadDecodesLineNotifyKey(t *testing.T) {
	// 64 hex chars → 32 bytes.
	hexKey := strings.Repeat("ab", 32)
	t.Setenv("LINE_NOTIFY_ENCRYPTION_KEY", hexKey)
	cfg := Load()
	if len(cfg.LineNotifyEncryptionKey) != 32 {
		t.Fatalf("expected 32-byte key, got %d bytes", len(cfg.LineNotifyEncryptionKey))
	}
}

func TestValidateLineNotify(t *testing.T) {
	cases := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{name: "valid 32-byte key", key: make([]byte, 32), wantErr: false},
		{name: "missing key", key: nil, wantErr: true},
		{name: "wrong length", key: make([]byte, 16), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{LineNotifyEncryptionKey: tc.key}
			err := cfg.ValidateLineNotify()
			if tc.wantErr && err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name      string
		jwtSecret string
		wantErr   bool
	}{
		{name: "valid 32+ char secret", jwtSecret: "0123456789012345678901234567890123", wantErr: false},
		{name: "missing secret", jwtSecret: "", wantErr: true},
		{name: "too short secret", jwtSecret: "short", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{JWTSecret: tc.jwtSecret}
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
