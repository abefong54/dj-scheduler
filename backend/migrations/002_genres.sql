-- backend/migrations/002_genres.sql
-- +goose Up
ALTER TABLE events ADD COLUMN IF NOT EXISTS genres TEXT[] DEFAULT '{}';
ALTER TABLE slots  ADD COLUMN IF NOT EXISTS genre  TEXT  DEFAULT '';

-- +goose Down
ALTER TABLE slots  DROP COLUMN IF EXISTS genre;
ALTER TABLE events DROP COLUMN IF EXISTS genres;
