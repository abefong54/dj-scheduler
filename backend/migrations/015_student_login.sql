-- backend/migrations/014_student_login.sql
-- Student Google login (ships first): a student IS a DJ row (is_student=true).
-- Add the two PII columns that let a student sign in with Google — email is set
-- by the teacher (the "invite"), google_id is bound on first Google login — and
-- the partial unique indexes that enforce "one Google identity = one student"
-- and a case-insensitive email lookup for the first-login auto-link.
-- email and google_id are PII: never logged, printed, or returned in public
-- payloads. Design: 2026-07-05-student-google-login-design.
-- +goose Up
ALTER TABLE djs ADD COLUMN email     TEXT;
ALTER TABLE djs ADD COLUMN google_id TEXT;

-- One Google identity = one student row (v1). Partial so many NULLs coexist.
CREATE UNIQUE INDEX idx_djs_google_id ON djs (google_id) WHERE google_id IS NOT NULL;
-- Fast, unique, case-insensitive email lookup for the first-login link.
CREATE UNIQUE INDEX idx_djs_email_lower ON djs (lower(email)) WHERE email IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_djs_email_lower;
DROP INDEX IF EXISTS idx_djs_google_id;
ALTER TABLE djs DROP COLUMN IF EXISTS google_id;
ALTER TABLE djs DROP COLUMN IF EXISTS email;
