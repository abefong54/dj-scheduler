package seed

import (
	"strings"
	"testing"
)

// TestDefaultCoreChallengesMatchSQL guards the two representations of the starter
// set against drift: the typed Go slice (used by a school-create hook) and the
// psql template (core_challenges.sql). Every Go entry must be well-formed and
// present in the SQL.
func TestDefaultCoreChallengesMatchSQL(t *testing.T) {
	if len(DefaultCoreChallenges) == 0 {
		t.Fatal("expected a non-empty default core challenge set")
	}
	seenOrder := map[int]bool{}
	for _, c := range DefaultCoreChallenges {
		if c.Title == "" || c.Description == "" || c.BadgeLabel == "" || c.BadgeIcon == "" {
			t.Fatalf("core challenge is missing a required field: %+v", c)
		}
		if seenOrder[c.SortOrder] {
			t.Fatalf("duplicate sort_order %d (ordering must be stable)", c.SortOrder)
		}
		seenOrder[c.SortOrder] = true
		if !strings.Contains(CoreChallengesSQL, c.Title) {
			t.Fatalf("challenge %q is in the Go slice but not in core_challenges.sql", c.Title)
		}
	}
}
