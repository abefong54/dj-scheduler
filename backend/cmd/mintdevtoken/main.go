// Command mintdevtoken prints a signed organizer JWT for local and E2E testing.
//
// It is a TEST affordance: it issues a token for an already-seeded organizer so
// automated tests (and developers) can authenticate without running the Google
// OAuth flow. The defaults match the E2E fixtures in backend/seed/seed_test.sql,
// so `JWT_SECRET=<test-secret> go run ./cmd/mintdevtoken` yields a token the
// test API will accept for the seeded organizer.
//
// This must never be wired into production: it mints a valid token for anyone
// who knows the signing secret, which is exactly why it lives behind a CLI fed
// the secret explicitly rather than an HTTP endpoint.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"eventlineup/internal/token"
)

func main() {
	organizerID := flag.String("organizer-id", "00000000-0000-0000-0000-000000000001",
		"organizer UUID to embed as the organizer_id claim (defaults to the seeded E2E organizer)")
	email := flag.String("email", "e2e-organizer@eventlineup.local", "organizer email claim")
	ttl := flag.Duration("ttl", 24*time.Hour, "token lifetime")
	secret := flag.String("secret", os.Getenv("JWT_SECRET"), "HMAC signing secret (defaults to $JWT_SECRET)")
	flag.Parse()

	// Mirror the backend's own minimum so a too-short secret fails here rather
	// than producing a token the API will reject.
	if len(*secret) < 32 {
		fmt.Fprintln(os.Stderr, "error: signing secret must be at least 32 characters (set -secret or $JWT_SECRET)")
		os.Exit(1)
	}

	tok, err := token.Sign(*secret, *organizerID, *email, *ttl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: sign token: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(tok)
}
