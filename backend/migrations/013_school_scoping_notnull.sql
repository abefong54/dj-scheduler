-- backend/migrations/013_school_scoping_notnull.sql
-- Scoping cutover (EL-104): school_id is now the tenant boundary for events + DJs,
-- so require it. DJs belong to the school (not a person), so organizer_id becomes
-- legacy/nullable there; events keep organizer_id as a created_by audit. The
-- full organizer_id retirement + rename is the deferred EL-102 cleanup.
-- +goose Up
ALTER TABLE events ALTER COLUMN school_id SET NOT NULL;
ALTER TABLE djs    ALTER COLUMN school_id SET NOT NULL;
ALTER TABLE djs    ALTER COLUMN organizer_id DROP NOT NULL;

-- +goose Down
-- Note: reversing organizer_id → NOT NULL fails if any DJ was created school-only
-- (organizer_id NULL) after this migration. Acceptable — Down is for fresh/CI DBs.
ALTER TABLE djs    ALTER COLUMN organizer_id SET NOT NULL;
ALTER TABLE djs    ALTER COLUMN school_id DROP NOT NULL;
ALTER TABLE events ALTER COLUMN school_id DROP NOT NULL;
