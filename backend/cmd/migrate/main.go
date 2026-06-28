package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := ensureMigrationsTable(ctx, pool); err != nil {
		log.Fatalf("ensure migrations table: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("glob migrations: %v", err)
	}
	sort.Strings(files)

	applied := 0
	for _, f := range files {
		name := filepath.Base(f)
		already, err := isApplied(ctx, pool, name)
		if err != nil {
			log.Fatalf("check %s: %v", name, err)
		}
		if already {
			fmt.Printf("  skip  %s\n", name)
			continue
		}

		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s: %v", name, err)
		}

		if err := apply(ctx, pool, name, string(sql)); err != nil {
			log.Fatalf("apply %s: %v", name, err)
		}
		fmt.Printf("  apply %s\n", name)
		applied++
	}

	if applied == 0 {
		fmt.Println("nothing to migrate")
	} else {
		fmt.Printf("%d migration(s) applied\n", applied)
	}
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT now()
		)`)
	return err
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`, name).
		Scan(&exists)
	return exists, err
}

func apply(ctx context.Context, pool *pgxpool.Pool, name, sql string) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
