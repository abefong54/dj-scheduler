// Package migrations holds the database schema migrations as goose SQL files,
// embedded into the binary so the runtime image needs no migrations/ directory
// on disk (the container ships only the compiled binaries).
package migrations

import "embed"

// FS holds every migration SQL file. goose reads versions from the numeric
// filename prefix (001, 002, …).
//
//go:embed *.sql
var FS embed.FS
