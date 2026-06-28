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
