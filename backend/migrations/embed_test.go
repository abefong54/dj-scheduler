package migrations

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// TestEmbeddedMigrationsParse verifies that every embedded .sql file is a
// well-formed goose migration and that all nine are present. It does not touch
// a database — sql.Open is lazy and goose reads the filesystem when the provider
// is constructed, so a bad -- +goose annotation or a missing file fails here
// without needing Postgres.
func TestEmbeddedMigrationsParse(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://user:pass@127.0.0.1:1/none")
	if err != nil {
		t.Fatalf("open placeholder db: %v", err)
	}
	defer db.Close()

	p, err := goose.NewProvider(goose.DialectPostgres, db, FS)
	if err != nil {
		t.Fatalf("goose could not parse embedded migrations: %v", err)
	}
	if got := len(p.ListSources()); got != 9 {
		t.Fatalf("embedded migrations = %d, want 9", got)
	}
}
