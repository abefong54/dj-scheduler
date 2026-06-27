package main

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInitDB(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := InitDB(url)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

// setupTestDB is shared by all handler tests in this package.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := InitDB(url)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
