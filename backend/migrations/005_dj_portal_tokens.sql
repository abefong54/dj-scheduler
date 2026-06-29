-- backend/migrations/005_dj_portal_tokens.sql
-- DJ self-service portal tokens (US-009). Each DJ can be issued a personal,
-- expiring portal link. Only the SHA-256 hash of the token is stored — the raw
-- token is shown to the organizer once and is never persisted.

ALTER TABLE djs ADD COLUMN IF NOT EXISTS portal_token_hash       TEXT;
ALTER TABLE djs ADD COLUMN IF NOT EXISTS portal_token_expires_at TIMESTAMPTZ;

-- Portal access looks DJs up by token hash, so index it. Partial index keeps it
-- small since most DJs will not have a token issued at any given time.
CREATE INDEX IF NOT EXISTS idx_djs_portal_token_hash
    ON djs (portal_token_hash) WHERE portal_token_hash IS NOT NULL;
