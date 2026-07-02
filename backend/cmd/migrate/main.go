// Command migrate applies the embedded goose migrations to the database named
// by DATABASE_URL.
//
// Usage:
//
//	migrate [up|down|status|version]
//
// The default subcommand is "up". Migrations are embedded in the binary
// (eventlineup/migrations), so no migrations/ directory is needed at runtime.
// A Postgres session advisory lock (goose session locker) makes concurrent
// "up" runs safe: exactly one instance applies, the rest are clean no-ops —
// which is what lets migrations run on every deploy across multiple instances.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"eventlineup/migrations"
)

func main() {
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()

	db, err := openDB(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	provider, err := newProvider(db)
	if err != nil {
		log.Fatalf("init migrations: %v", err)
	}

	if err := run(ctx, provider, cmd, os.Stdout); err != nil {
		log.Fatalf("%s: %v", cmd, err)
	}
}

// openDB opens a database/sql handle over pgx's stdlib driver — goose speaks
// database/sql, while the app itself uses a pgxpool.
func openDB(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	// A one-shot migrator needs only a couple of connections. Keep the cap at or
	// above 2: the goose session locker pins one connection to hold the advisory
	// lock while migrations run, so a max of 1 would self-deadlock.
	db.SetMaxOpenConns(4)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Bound the initial connectivity check so a bad DATABASE_URL fails fast here
	// rather than hanging, or surfacing mid-migration.
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// newProvider builds a goose provider over the embedded migrations, guarded by a
// Postgres session advisory lock so concurrent runs don't race.
func newProvider(db *sql.DB) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, err
	}
	return goose.NewProvider(goose.DialectPostgres, db, migrations.FS,
		goose.WithSessionLocker(locker))
}

// run dispatches a subcommand against the provider and reports what it did.
func run(ctx context.Context, p *goose.Provider, cmd string, out io.Writer) error {
	switch cmd {
	case "up":
		results, err := p.Up(ctx)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Fprintln(out, "nothing to migrate")
			return nil
		}
		for _, r := range results {
			fmt.Fprintf(out, "  applied  %s\n", r.Source.Path)
		}
		fmt.Fprintf(out, "%d migration(s) applied\n", len(results))
		return nil

	case "down":
		r, err := p.Down(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "  reverted  %s\n", r.Source.Path)
		return nil

	case "status":
		st, err := p.Status(ctx)
		if err != nil {
			return err
		}
		for _, s := range st {
			state := "pending"
			if s.State == goose.StateApplied {
				state = "applied"
			}
			fmt.Fprintf(out, "  %-8s %s\n", state, s.Source.Path)
		}
		return nil

	case "version":
		v, err := p.GetDBVersion(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "current version: %d\n", v)
		return nil

	default:
		return fmt.Errorf("unknown command %q (want up|down|status|version)", cmd)
	}
}
