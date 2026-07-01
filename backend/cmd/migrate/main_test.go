package main

import (
	"context"
	"os"
	"testing"
)

// TestMigrateUpDownIdempotent exercises the real goose flow against a throwaway
// Postgres named by GOOSE_TEST_DATABASE_URL (a FRESH database — it migrates from
// version 0). It is skipped in the normal unit run; the CI migration job and
// local `docker compose` supply the URL. It asserts the four behaviours the
// ticket calls out: up applies everything, a second up is a no-op (idempotent),
// down reverses the latest, and up re-applies it.
func TestMigrateUpDownIdempotent(t *testing.T) {
	url := os.Getenv("GOOSE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("GOOSE_TEST_DATABASE_URL not set")
	}

	db, err := openDB(context.Background(), url)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	p, err := newProvider(db)
	if err != nil {
		t.Fatalf("newProvider: %v", err)
	}
	ctx := context.Background()

	if _, err := p.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	if v, _ := p.GetDBVersion(ctx); v != 9 {
		t.Fatalf("version after up = %d, want 9", v)
	}

	// Idempotent: a second up on an up-to-date DB applies nothing.
	res, err := p.Up(ctx)
	if err != nil {
		t.Fatalf("second up: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("second up applied %d migration(s), want 0", len(res))
	}

	// Down reverses exactly the latest migration.
	if _, err := p.Down(ctx); err != nil {
		t.Fatalf("down: %v", err)
	}
	if v, _ := p.GetDBVersion(ctx); v != 8 {
		t.Fatalf("version after down = %d, want 8", v)
	}

	// And it re-applies cleanly.
	if _, err := p.Up(ctx); err != nil {
		t.Fatalf("re-up: %v", err)
	}
	if v, _ := p.GetDBVersion(ctx); v != 9 {
		t.Fatalf("version after re-up = %d, want 9", v)
	}
}
