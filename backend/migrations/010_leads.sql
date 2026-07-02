-- backend/migrations/010_leads.sql
-- Demo/contact leads captured from the public Soundcheck landing page (EL-084).
-- `organization` (not `school`) so the model isn't locked to the DJ-school beachhead —
-- a studio, crew, or company is just as valid a lead.
-- +goose Up
CREATE TABLE IF NOT EXISTS leads (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name         TEXT NOT NULL,
  organization TEXT NOT NULL DEFAULT '',
  email        TEXT NOT NULL,
  message      TEXT NOT NULL DEFAULT '',
  source       TEXT NOT NULL DEFAULT 'landing',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS leads;
