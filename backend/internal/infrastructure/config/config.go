package config

import (
	"encoding/hex"
	"errors"
	"os"
)

// minJWTSecretLen is the minimum acceptable length for the JWT signing secret.
const minJWTSecretLen = 32

// lineNotifyKeyLen is the required AES key length for encrypting LINE Notify
// tokens at rest (256-bit AES-GCM).
const lineNotifyKeyLen = 32

type Config struct {
	DatabaseURL string
	Port        string
	FrontendURL string
	JWTSecret   string

	// SecureCookies marks auth cookies Secure (HTTPS-only). Off by default for
	// local http dev; set COOKIE_SECURE=true in production.
	SecureCookies bool

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// LINE Login (organizer sign-in via LINE, EL-110). Optional: when all three
	// are present the /auth/line routes + button turn on; all-empty disables it.
	// Distinct from LINE Notify (LineNotifyEncryptionKey) below.
	LineChannelID     string
	LineChannelSecret string
	LineRedirectURL   string

	// LineNotifyEncryptionKey is the AES-256 key (decoded from the hex env var
	// LINE_NOTIFY_ENCRYPTION_KEY) used to encrypt per-event LINE Notify tokens
	// at rest. nil if unset or not valid hex; ValidateLineNotify rejects that.
	LineNotifyEncryptionKey []byte
}

func Load() Config {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        port,
		FrontendURL: frontendURL,
		JWTSecret:   os.Getenv("JWT_SECRET"),

		SecureCookies: os.Getenv("COOKIE_SECURE") == "true",

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),

		LineChannelID:     os.Getenv("LINE_CHANNEL_ID"),
		LineChannelSecret: os.Getenv("LINE_CHANNEL_SECRET"),
		LineRedirectURL:   os.Getenv("LINE_REDIRECT_URL"),

		// Invalid hex decodes to nil here; ValidateLineNotify reports it.
		LineNotifyEncryptionKey: decodeHexKey(os.Getenv("LINE_NOTIFY_ENCRYPTION_KEY")),
	}
}

// decodeHexKey decodes a hex string to bytes, returning nil on empty or invalid
// input so the caller fails validation rather than starting with a bad key.
func decodeHexKey(s string) []byte {
	if s == "" {
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}

// Validate checks that required configuration is present and well-formed.
// It is meant to be called at startup so the server fails fast on misconfiguration.
func (c Config) Validate() error {
	if len(c.JWTSecret) < minJWTSecretLen {
		return errors.New("JWT_SECRET must be set and at least 32 characters")
	}
	return nil
}

// ValidateLineNotify checks that the LINE Notify token-encryption key is a
// well-formed AES-256 key. Called at startup so the server refuses to run with
// a missing or malformed LINE_NOTIFY_ENCRYPTION_KEY.
func (c Config) ValidateLineNotify() error {
	if len(c.LineNotifyEncryptionKey) != lineNotifyKeyLen {
		return errors.New("LINE_NOTIFY_ENCRYPTION_KEY must be a 64-char hex string (32-byte AES-256 key)")
	}
	return nil
}

// LineLoginConfigured reports whether LINE Login is fully configured. LINE login
// is optional — all-empty means "disabled" (Google stays required).
func (c Config) LineLoginConfigured() bool {
	return c.LineChannelID != "" && c.LineChannelSecret != "" && c.LineRedirectURL != ""
}

// ValidateLine rejects a PARTIAL LINE Login config (some but not all settings)
// so a typo can't silently disable the button. Fully set or fully empty pass.
func (c Config) ValidateLine() error {
	n := 0
	for _, v := range []string{c.LineChannelID, c.LineChannelSecret, c.LineRedirectURL} {
		if v != "" {
			n++
		}
	}
	if n != 0 && n != 3 {
		return errors.New("LINE Login partially configured — set all of LINE_CHANNEL_ID, LINE_CHANNEL_SECRET, LINE_REDIRECT_URL, or none")
	}
	return nil
}

// ValidateGoogle checks that the Google SSO configuration is fully present.
func (c Config) ValidateGoogle() error {
	if c.GoogleClientID == "" || c.GoogleClientSecret == "" || c.GoogleRedirectURL == "" {
		return errors.New("GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL must all be set")
	}
	return nil
}
