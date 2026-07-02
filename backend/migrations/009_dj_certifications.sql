-- backend/migrations/009_dj_certifications.sql
-- DJ certifications (EL-019). Which genres each DJ is cleared to perform, and
-- whether they're an active student (gate applies) or a graduate/pro (gate
-- bypassed). Existing DJs default to no certifications and student status.
-- +goose Up
ALTER TABLE djs ADD COLUMN IF NOT EXISTS certifications TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE djs ADD COLUMN IF NOT EXISTS is_student BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE djs DROP COLUMN IF EXISTS is_student;
ALTER TABLE djs DROP COLUMN IF EXISTS certifications;
