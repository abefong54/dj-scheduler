package database

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func deadlockErr() error { return &pgconn.PgError{Code: "40P01"} }

func TestRetryOnDeadlock_SucceedsFirstTry(t *testing.T) {
	calls := 0
	err := retryOnDeadlock(func() error { calls++; return nil })
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call, got %d", calls)
	}
}

// A deadlock is transient: once the winner commits, the retry cleanly loses the
// overlap race, so retryOnDeadlock must keep going until a non-deadlock outcome.
func TestRetryOnDeadlock_RetriesUntilSuccess(t *testing.T) {
	calls := 0
	err := retryOnDeadlock(func() error {
		calls++
		if calls < 3 {
			return deadlockErr()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

// Only deadlocks are transient. A 23P01 exclusion violation is the definitive
// "you lost the race" answer and must be returned untouched on the first try.
func TestRetryOnDeadlock_DoesNotRetryNonDeadlock(t *testing.T) {
	calls := 0
	sentinel := &pgconn.PgError{Code: "23P01"}
	err := retryOnDeadlock(func() error { calls++; return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the non-deadlock error returned unchanged, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("a non-deadlock error must not be retried; got %d calls", calls)
	}
}

func TestRetryOnDeadlock_GivesUpAfterMaxAndReturnsLastError(t *testing.T) {
	calls := 0
	err := retryOnDeadlock(func() error { calls++; return deadlockErr() })
	if !isDeadlock(err) {
		t.Fatalf("expected the last deadlock error after exhausting retries, got %v", err)
	}
	if calls != maxWriteRetries {
		t.Fatalf("expected %d attempts, got %d", maxWriteRetries, calls)
	}
}
