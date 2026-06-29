-- backend/migrations/008_line_notify.sql
-- Per-event LINE Notify token (US-006). The token is encrypted with AES-256-GCM
-- before storage; NULL means LINE Notify is disabled for the event. The raw
-- token is never stored or returned by the API.

ALTER TABLE events ADD COLUMN IF NOT EXISTS line_notify_token_enc TEXT;
