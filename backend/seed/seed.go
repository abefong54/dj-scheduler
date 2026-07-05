// Package seed holds curated starter data that new schools begin with. Today
// that is the default set of CORE challenges (a skill ladder every school gets),
// kept both as a psql-runnable template (core_challenges.sql) and as a typed Go
// slice so a future school-create hook can insert it with a parameterized query
// instead of shelling out to psql. A test keeps the two representations in sync.
package seed

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgconn"
)

// CoreChallengesSQL is the psql-runnable template for the default core
// challenges (see the file header for the :school_id / :created_by variables).
// It is the human/source-of-truth companion to DefaultCoreChallenges below.
//
//go:embed core_challenges.sql
var CoreChallengesSQL string

// CoreChallenge is one entry in the default starter set: a school-wide core
// challenge that awards a badge when a student passes it (migration 015).
type CoreChallenge struct {
	Title       string
	Description string
	Skill       string
	BadgeLabel  string
	BadgeIcon   string
	SortOrder   int
}

// DefaultCoreChallenges is the curated skill ladder every new school starts
// with, ordered easiest -> hardest. It mirrors core_challenges.sql exactly
// (TestDefaultCoreChallengesMatchSQL guards the two against drift).
var DefaultCoreChallenges = []CoreChallenge{
	{"First Blend", "Mix smoothly from one track into another so the transition sounds like a single continuous groove.", "mixing", "First Blend", "🎧", 10},
	{"Beatmatcher", "Beatmatch two tracks by ear and hold them locked for at least 16 bars — no sync button.", "beatmatching", "Beatmatcher", "🥁", 20},
	{"Clean Cut", "Drop a crisp, on-beat cut between two tracks with no train-wreck and no dead air.", "transitions", "Clean Cut", "✂️", 30},
	{"Genre Bender", "Blend two tracks from clearly different genres or tempos into one coherent transition.", "mixing", "Genre Bender", "🎚️", 40},
	{"Loop Master", "Use a loop or beat-roll musically to extend a phrase and set up the next drop.", "performance", "Loop Master", "🔁", 50},
	{"Crowd Reader", "Read the room and adapt your next selection live to keep the energy moving.", "selection", "Crowd Reader", "👀", 60},
	{"Headliner", "Perform a complete 30-minute set with clean transitions and a deliberate energy arc.", "performance", "Headliner", "🏆", 70},
}

// Execer is the slice of pgx a seed insert needs: a single parameterized Exec.
// Both *pgxpool.Pool and pgx.Tx satisfy it, so callers can seed inside or
// outside a transaction.
type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// insertCoreChallenge inserts one default core challenge for the given school,
// skipping when a same-titled challenge already exists so the starter set can be
// (re)applied idempotently. Mirrors the WHERE NOT EXISTS in core_challenges.sql.
const insertCoreChallenge = `
INSERT INTO challenges (school_id, type, title, description, skill, badge_label, badge_icon, sort_order, created_by)
SELECT $1, 'core', $2, $3, $4, $5, $6, $7, $8
WHERE NOT EXISTS (
  SELECT 1 FROM challenges c WHERE c.school_id = $1 AND c.title = $2
)`

// InsertDefaultCoreChallenges seeds the curated starter set for schoolID,
// authored by createdBy (an organizer id). It is idempotent per school and
// returns the number of challenges actually inserted. Intended for a
// school-create hook to call; safe to run against an already-seeded school.
func InsertDefaultCoreChallenges(ctx context.Context, db Execer, schoolID, createdBy string) (int, error) {
	inserted := 0
	for _, c := range DefaultCoreChallenges {
		tag, err := db.Exec(ctx, insertCoreChallenge,
			schoolID, c.Title, c.Description, c.Skill, c.BadgeLabel, c.BadgeIcon, c.SortOrder, createdBy)
		if err != nil {
			return inserted, err
		}
		inserted += int(tag.RowsAffected())
	}
	return inserted, nil
}
