-- backend/migrations/006_slot_confirmation.sql
-- DJ self-service confirmation (US-011). A DJ can confirm or flag each of their
-- slots from the portal. NULL means no response yet.
-- +goose Up
ALTER TABLE slots ADD COLUMN IF NOT EXISTS dj_confirmation TEXT
    CHECK (dj_confirmation IN ('confirmed', 'flagged'));

-- +goose Down
-- Dropping the column also drops its CHECK constraint.
ALTER TABLE slots DROP COLUMN IF EXISTS dj_confirmation;
