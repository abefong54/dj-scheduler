package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

// eventOwnedBy reports whether the event exists and belongs to the organizer.
// Used by the stage/slot list paths to tell an empty-but-owned event (→ 200 with
// an empty list) apart from one that isn't the organizer's (→ ErrNotFound).
func eventOwnedBy(ctx context.Context, pool *pgxpool.Pool, eventID, organizerID string) (bool, error) {
	var owned bool
	err := pool.QueryRow(ctx, queryEventOwned, eventID, organizerID).Scan(&owned)
	return owned, err
}
