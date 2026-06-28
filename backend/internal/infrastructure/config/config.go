package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	FrontendURL string
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
	}
}
