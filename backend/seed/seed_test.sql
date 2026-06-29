-- backend/seed/seed_test.sql
--
-- Deterministic fixtures for the Playwright E2E suite. Re-runnable: it truncates
-- the mutable tables first, so the SAME file doubles as the between-test reset
-- (see `make e2e-seed` / `make e2e-reset`).
--
-- TEST DATA ONLY. This targets the throwaway database in docker-compose.test.yml
-- and must never be run against a real database. The values below (organizer,
-- portal token, secret-derived IDs) are fabricated fixtures, not real PII.
--
-- Fixed UUIDs let the test harness reference rows without a lookup round-trip:
--   organizer  00000000-0000-0000-0000-000000000001
--   dj1        00000000-0000-0000-0000-0000000000d1  (has a portal token)
--   dj2        00000000-0000-0000-0000-0000000000d2
--   event      00000000-0000-0000-0000-0000000000e1
--   stage1/2   ...05a1 / ...05a2
--   slot1/2    ...005101 / ...005102

BEGIN;

-- Wipe everything the suite mutates. CASCADE clears stages/slots via their FKs.
TRUNCATE slots, stages, events, djs, organizers RESTART IDENTITY CASCADE;

-- Test organizer. Its id MUST match the organizer_id claim in the minted JWT
-- (`make e2e-token` / cmd/mintdevtoken default), or scoped queries return empty.
INSERT INTO organizers (id, email, name, google_id) VALUES
  ('00000000-0000-0000-0000-000000000001', 'e2e-organizer@eventlineup.local', 'E2E Organizer', 'e2e-google-id-1');

-- DJs. dj1 carries a portal token whose RAW value is 'e2e-portal-token-dj1'.
-- Only the SHA-256 hash is stored, computed exactly as the app does
-- (hex-encoded sha256, no salt — see internal/usecase/dj/portal.go hashToken).
-- The harness visits /dj/portal?token=e2e-portal-token-dj1.
INSERT INTO djs (id, organizer_id, name, genre_tags, portal_token_hash, portal_token_expires_at) VALUES
  ('00000000-0000-0000-0000-0000000000d1', '00000000-0000-0000-0000-000000000001', 'DJ Testa', ARRAY['house','techno'],
     encode(digest('e2e-portal-token-dj1', 'sha256'), 'hex'), now() + interval '365 days'),
  ('00000000-0000-0000-0000-0000000000d2', '00000000-0000-0000-0000-000000000001', 'DJ Beta', ARRAY['ambient'],
     NULL, NULL);

-- One event with two stages.
INSERT INTO events (id, organizer_id, name, venue_name, start_date, end_date, genres) VALUES
  ('00000000-0000-0000-0000-0000000000e1', '00000000-0000-0000-0000-000000000001', 'E2E Fest', 'Test Arena',
     DATE '2026-07-01', DATE '2026-07-02', ARRAY['house','techno','ambient']);

INSERT INTO stages (id, event_id, name, color, display_order) VALUES
  ('00000000-0000-0000-0000-0000000005a1', '00000000-0000-0000-0000-0000000000e1', 'Main Stage', '#6366F1', 0),
  ('00000000-0000-0000-0000-0000000005a2', '00000000-0000-0000-0000-0000000000e1', 'Side Stage', '#10B981', 1);

-- Two slots, one per DJ, both awaiting confirmation (dj_confirmation NULL).
INSERT INTO slots (id, event_id, stage_id, dj_id, slot_date, start_time, end_time, notes, genre, dj_confirmation) VALUES
  ('00000000-0000-0000-0000-000000005101', '00000000-0000-0000-0000-0000000000e1', '00000000-0000-0000-0000-0000000005a1',
     '00000000-0000-0000-0000-0000000000d1', DATE '2026-07-01', TIME '20:00', TIME '21:00', '', 'house', NULL),
  ('00000000-0000-0000-0000-000000005102', '00000000-0000-0000-0000-0000000000e1', '00000000-0000-0000-0000-0000000005a2',
     '00000000-0000-0000-0000-0000000000d2', DATE '2026-07-01', TIME '21:00', TIME '22:00', '', 'ambient', NULL);

COMMIT;
