-- backend/migrations/012_school_scoping.sql
-- Backfill (EL-103): every existing organizer becomes the owner of a personal
-- school, and their events + DJs are stamped with that school_id. school_id is
-- kept NULLABLE for now — this migration is additive and non-breaking; the
-- NOT NULL constraint, the organizer_id retirement on djs, and the scoping
-- cutover land together in EL-104. Design: 2026-07-04-multi-tenant-schools.
-- +goose Up
INSERT INTO schools (name, created_by)
  SELECT o.name || '''s School', o.id FROM organizers o;

INSERT INTO memberships (school_id, organizer_id, email, role, status)
  SELECT s.id, o.id, o.email, 'owner', 'active'
  FROM organizers o JOIN schools s ON s.created_by = o.id;

ALTER TABLE events ADD COLUMN IF NOT EXISTS school_id UUID REFERENCES schools(id);
ALTER TABLE djs    ADD COLUMN IF NOT EXISTS school_id UUID REFERENCES schools(id);

UPDATE events e SET school_id = s.id FROM schools s
  WHERE s.created_by = e.organizer_id AND e.school_id IS NULL;
UPDATE djs d SET school_id = s.id FROM schools s
  WHERE s.created_by = d.organizer_id AND d.school_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_events_school ON events(school_id);
CREATE INDEX IF NOT EXISTS idx_djs_school ON djs(school_id);

-- +goose Down
DROP INDEX IF EXISTS idx_djs_school;
DROP INDEX IF EXISTS idx_events_school;
ALTER TABLE djs    DROP COLUMN IF EXISTS school_id;
ALTER TABLE events DROP COLUMN IF EXISTS school_id;
DELETE FROM memberships;
DELETE FROM schools;
