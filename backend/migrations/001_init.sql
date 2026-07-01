-- backend/migrations/001_init.sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS djs (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  genre_tags TEXT[] DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS events (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  venue_name  TEXT NOT NULL,
  start_date  DATE NOT NULL,
  end_date    DATE NOT NULL,
  created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS stages (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id      UUID REFERENCES events(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  color         TEXT NOT NULL DEFAULT '#6366F1',
  display_order INTEGER DEFAULT 0,
  created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS slots (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   UUID REFERENCES events(id) ON DELETE CASCADE,
  stage_id   UUID REFERENCES stages(id) ON DELETE CASCADE,
  dj_id      UUID REFERENCES djs(id) ON DELETE SET NULL,
  slot_date  DATE NOT NULL,
  start_time TIME NOT NULL,
  end_time   TIME NOT NULL,
  notes      TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stages_event ON stages(event_id);
CREATE INDEX IF NOT EXISTS idx_slots_event  ON slots(event_id);
CREATE INDEX IF NOT EXISTS idx_slots_date   ON slots(slot_date);

-- +goose Down
-- Drop in reverse dependency order (slots → stages → events → djs). The indexes
-- are dropped implicitly with their tables. pgcrypto is left installed — it may
-- be shared by other databases on the instance and is harmless to keep.
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS stages;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS djs;
