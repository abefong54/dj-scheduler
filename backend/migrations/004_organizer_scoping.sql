-- backend/migrations/004_organizer_scoping.sql
-- Scope events and djs to an organizer. Non-destructive: existing rows are
-- backfilled to a synthetic "legacy" organizer before NOT NULL is enforced, so
-- pre-auth data is preserved rather than dropped.

-- 1. Add the column nullable so existing rows survive.
ALTER TABLE events ADD COLUMN IF NOT EXISTS organizer_id UUID REFERENCES organizers(id);
ALTER TABLE djs   ADD COLUMN IF NOT EXISTS organizer_id UUID REFERENCES organizers(id);

-- 2. Ensure a legacy organizer exists to own pre-auth data.
INSERT INTO organizers (email, name, google_id)
VALUES ('legacy@eventlineup.local', 'Legacy (pre-auth)', 'legacy-pre-auth')
ON CONFLICT (google_id) DO NOTHING;

-- 3. Backfill any unscoped rows to the legacy organizer.
UPDATE events
SET organizer_id = (SELECT id FROM organizers WHERE google_id = 'legacy-pre-auth')
WHERE organizer_id IS NULL;

UPDATE djs
SET organizer_id = (SELECT id FROM organizers WHERE google_id = 'legacy-pre-auth')
WHERE organizer_id IS NULL;

-- 4. Now that every row is scoped, enforce NOT NULL.
ALTER TABLE events ALTER COLUMN organizer_id SET NOT NULL;
ALTER TABLE djs   ALTER COLUMN organizer_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_organizer ON events(organizer_id);
CREATE INDEX IF NOT EXISTS idx_djs_organizer    ON djs(organizer_id);
