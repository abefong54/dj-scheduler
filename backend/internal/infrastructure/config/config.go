package config

import (
	"errors"
	"os"
)

// minJWTSecretLen is the minimum acceptable length for the JWT signing secret.
const minJWTSecretLen = 32

type Config struct {
	DatabaseURL string
	Port        string
	FrontendURL string
	JWTSecret   string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
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

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}
}

// Validate checks that required configuration is present and well-formed.
// It is meant to be called at startup so the server fails fast on misconfiguration.
func (c Config) Validate() error {
	if len(c.JWTSecret) < minJWTSecretLen {
		return errors.New("JWT_SECRET must be set and at least 32 characters")
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
