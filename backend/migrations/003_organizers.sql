-- backend/migrations/003_organizers.sql
-- Organizer accounts, created on first Google sign-in (US-001).
-- +goose Up
CREATE TABLE IF NOT EXISTS organizers (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email      TEXT NOT NULL UNIQUE,
  name       TEXT NOT NULL,
  google_id  TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS organizers;
