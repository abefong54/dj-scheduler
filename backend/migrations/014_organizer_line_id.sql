-- backend/migrations/014_organizer_line_id.sql
-- LINE login (EL-109): organizers can now authenticate via a second provider.
-- Add a nullable line_id, relax google_id to nullable (a LINE-first user has no
-- Google id), and require at least one provider identity. email stays the
-- linking key (NOT NULL UNIQUE), so a Google + LINE sign-in with the same email
-- resolves to one account.
-- +goose Up
ALTER TABLE organizers ADD COLUMN line_id TEXT UNIQUE;
ALTER TABLE organizers ALTER COLUMN google_id DROP NOT NULL;
ALTER TABLE organizers ADD CONSTRAINT organizers_provider_present
  CHECK (google_id IS NOT NULL OR line_id IS NOT NULL);

-- +goose Down
-- Reversing google_id → NOT NULL fails if any LINE-only organizer (google_id
-- NULL) was created after this migration. Acceptable — Down is for fresh/CI DBs.
ALTER TABLE organizers DROP CONSTRAINT organizers_provider_present;
ALTER TABLE organizers ALTER COLUMN google_id SET NOT NULL;
ALTER TABLE organizers DROP COLUMN line_id;
