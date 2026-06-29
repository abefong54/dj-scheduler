package database_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/infrastructure/database"
	slotuc "eventlineup/internal/usecase/slot"
)

// seedEventStage inserts an organizer, event, and stage, returning the event,
// stage, and organizer IDs and registering cleanup.
func seedEventStage(t *testing.T, pool *pgxpool.Pool) (eventID, stageID, orgID string) {
	t.Helper()
	ctx := context.Background()
	if err := pool.QueryRow(ctx,
		`INSERT INTO organizers (email, name, google_id)
		 VALUES (gen_random_uuid()::text||'@t.local', 't', gen_random_uuid()::text)
		 RETURNING id`).Scan(&orgID); err != nil {
		t.Fatalf("insert organizer: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM events WHERE organizer_id = $1", orgID)
		pool.Exec(ctx, "DELETE FROM organizers WHERE id = $1", orgID)
	})
	if err := pool.QueryRow(ctx,
		`INSERT INTO events (name, venue_name, start_date, end_date, organizer_id)
		 VALUES ('e', 'v', '2026-07-25', '2026-07-26', $1) RETURNING id`, orgID).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO stages (event_id, name, color) VALUES ($1, 's', '#ffffff') RETURNING id`,
		eventID).Scan(&stageID); err != nil {
		t.Fatalf("insert stage: %v", err)
	}
	return eventID, stageID, orgID
}

func TestSlotRepo_RejectsOverlapAtDatabaseLevel(t *testing.T) {
	pool := orgTestPool(t)
	ctx := context.Background()
	repo := database.NewSlotRepository(pool)
	eventID, stageID, orgID := seedEventStage(t, pool)

	// First slot: 16:00–17:30.
	if _, err := repo.Create(ctx, model.Slot{
		StageID: stageID, SlotDate: "2026-07-25", StartTime: "16:00", EndTime: "17:30",
	}, eventID, orgID); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Overlapping slot on the same stage. The repository runs no conflict check of
	// its own (that lives in the use case), so only the database constraint can
	// stop this — exactly the guarantee that protects against the check-then-write
	// race.
	_, err := repo.Create(ctx, model.Slot{
		StageID: stageID, SlotDate: "2026-07-25", StartTime: "17:00", EndTime: "18:00",
	}, eventID, orgID)
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected ErrConflict from the DB overlap constraint, got %v", err)
	}
}

// TestSlotUseCase_ConcurrentOverlapsOnlyOneWins fires two identical (definitely
// overlapping) slot creates at once through the use case. With only the
// application-level check this could let both through; the database exclusion
// constraint guarantees exactly one commits and the other gets a conflict.
func TestSlotUseCase_ConcurrentOverlapsOnlyOneWins(t *testing.T) {
	pool := orgTestPool(t)
	ctx := context.Background()
	uc := slotuc.New(database.NewSlotRepository(pool))
	eventID, stageID, orgID := seedEventStage(t, pool)

	const n = 2
	var wg sync.WaitGroup
	results := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = uc.Create(ctx, model.Slot{
				StageID: stageID, SlotDate: "2026-07-25", StartTime: "20:00", EndTime: "21:00",
			}, eventID, orgID)
		}(i)
	}
	wg.Wait()

	success, conflict := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, apperrors.ErrConflict):
			conflict++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("expected exactly one success and one conflict, got success=%d conflict=%d", success, conflict)
	}
}
