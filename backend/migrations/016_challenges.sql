-- backend/migrations/015_challenges.sql
-- Student Challenges & Badges. Two new tables carry the whole feature:
--   challenges  — teacher-authored tasks (core = school-wide + badge; custom =
--                 assigned to one student, no badge).
--   submissions — a student's attempt at a challenge: simultaneously the feed
--                 item and the progress driver. One active row per
--                 (dj_id, challenge_id). When a core challenge is passed, the
--                 challenge's badge_label/badge_icon are SNAPSHOTTED onto the
--                 submission (earned_badge_*) so the earned badge is permanent.
-- Both tables are school-scoped (school_id NOT NULL, indexed) exactly like
-- migrations 012/013. Design: 2026-07-05-student-challenges-design.
-- +goose Up
CREATE TABLE challenges (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    type           TEXT NOT NULL,                 -- 'core' | 'custom'
    title          TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    skill          TEXT,                          -- optional tag
    assigned_dj_id UUID REFERENCES djs(id),       -- custom only, else NULL
    badge_label    TEXT,                          -- core only
    badge_icon     TEXT,                          -- core only
    sort_order     INT  NOT NULL DEFAULT 0,
    created_by     UUID NOT NULL,                 -- organizer
    archived       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE submissions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id          UUID NOT NULL REFERENCES schools(id),
    challenge_id       UUID NOT NULL REFERENCES challenges(id),
    dj_id              UUID NOT NULL REFERENCES djs(id),
    media_kind         TEXT NOT NULL,             -- 'embed' (future: 'file')
    media_provider     TEXT NOT NULL,             -- youtube | soundcloud | mixcloud
    media_ref          TEXT NOT NULL,             -- extracted id / normalized url
    caption            TEXT NOT NULL DEFAULT '',
    status             TEXT NOT NULL,             -- 'posted' | 'passed' | 'needs_work'
    review_note        TEXT,                      -- nullable
    earned_badge_label TEXT,                      -- snapshot at pass, core only
    earned_badge_icon  TEXT,                      -- snapshot at pass, core only
    reviewed_by        UUID,                      -- nullable organizer
    reviewed_at        TIMESTAMPTZ,               -- nullable
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- One active submission per student per challenge; re-work updates the row.
    UNIQUE (dj_id, challenge_id)
);

CREATE INDEX idx_challenges_school ON challenges(school_id);
CREATE INDEX idx_submissions_school ON submissions(school_id);

-- +goose Down
DROP INDEX IF EXISTS idx_submissions_school;
DROP INDEX IF EXISTS idx_challenges_school;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS challenges;
