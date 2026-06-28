package handler_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/infrastructure/database"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := database.InitDB(url)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func createTestEvent(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO events (name, venue_name, start_date, end_date)
		 VALUES ('test-event', 'Test Venue', '2026-07-25', '2026-07-26')
		 RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("createTestEvent: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", id)
	})
	return id
}

func createTestStage(t *testing.T, pool *pgxpool.Pool, eventID string) string {
	t.Helper()
	var id string
	pool.QueryRow(context.Background(),
		`INSERT INTO stages (event_id, name, color) VALUES ($1, 'Test Stage', '#6366F1') RETURNING id`,
		eventID).Scan(&id)
	return id
}

func createTestSlot(t *testing.T, pool *pgxpool.Pool, eventID, stageID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO slots (event_id, stage_id, slot_date, start_time, end_time)
		 VALUES ($1, $2, '2026-07-25', '16:00', '17:30') RETURNING id`,
		eventID, stageID).Scan(&id)
	if err != nil {
		t.Fatalf("createTestSlot: %v", err)
	}
	return id
}
