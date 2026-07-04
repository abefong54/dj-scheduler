-- backend/migrations/011_schools_and_memberships.sql
-- Multi-tenant schools (EL-103). A school is the tenant that owns events + DJs;
-- organizers join a school through memberships. Additive + non-breaking: the
-- existing organizer scoping keeps working alongside the new school_id (added and
-- backfilled in 012). The NOT NULL + organizer_id retirement land later with the
-- scoping cutover (EL-104). Design: 2026-07-04-multi-tenant-schools.
-- +goose Up
CREATE TABLE IF NOT EXISTS schools (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  city       TEXT NOT NULL DEFAULT '',
  created_by UUID REFERENCES organizers(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS memberships (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id    UUID NOT NULL REFERENCES schools(id),
  organizer_id UUID REFERENCES organizers(id),   -- NULL while an invite is pending
  email        TEXT NOT NULL,
  role         TEXT NOT NULL DEFAULT 'member',    -- owner | admin | member
  status       TEXT NOT NULL DEFAULT 'pending',   -- pending | active
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (school_id, email)
);
CREATE INDEX IF NOT EXISTS idx_memberships_organizer ON memberships(organizer_id);
CREATE INDEX IF NOT EXISTS idx_memberships_email ON memberships(email);

-- +goose Down
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS schools;
