package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(connStr string) (*pgxpool.Pool, error) {
	cfg, err := poolConfig(connStr)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

// poolConfig parses the connection string and tunes the pgx pool. Explicit
// sizing matters on a managed/serverless Postgres (Neon): MinConns keeps a warm
// floor of connections so a request doesn't pay a fresh TLS + auth handshake
// after an idle gap, and MaxConns caps concurrency well under the server limit.
// Lifetimes recycle connections so a stale one (e.g. dropped server-side) is
// pruned rather than handed to a query.
func poolConfig(connStr string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.HealthCheckPeriod = time.Minute
	return cfg, nil
}

// eventOwnedBy reports whether the event exists and belongs to the organizer.
// Used by the stage/slot list paths to tell an empty-but-owned event (→ 200 with
// an empty list) apart from one that isn't the organizer's (→ ErrNotFound).
func eventOwnedBy(ctx context.Context, pool *pgxpool.Pool, eventID, organizerID string) (bool, error) {
	var owned bool
	err := pool.QueryRow(ctx, queryEventOwned, eventID, organizerID).Scan(&owned)
	return owned, err
}
