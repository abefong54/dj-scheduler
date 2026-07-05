-- backend/seed/core_challenges.sql
--
-- The curated set of default CORE challenges every new school starts with, so
-- teachers see a ready-made skill ladder instead of a blank slate. Each is a
-- core challenge (school-wide + earns a badge) covering a foundational DJ skill,
-- ordered easiest -> hardest by sort_order.
--
-- This file is the single source of truth for the default set. It is a template:
-- it references the target school via the psql variables :school_id and
-- :created_by (an organizer id), so it is runnable directly with psql, e.g.
--
--   psql "$DATABASE_URL" \
--     -v school_id="'<uuid>'" -v created_by="'<organizer-uuid>'" \
--     -f backend/seed/core_challenges.sql
--
-- The Go seed package embeds this file (see embed.go) and substitutes the two
-- placeholders, so future school-create wiring can reuse the exact same list.
--
-- Idempotent per school: skips challenges whose (school_id, title) already
-- exists, so re-running never duplicates the starter set.

INSERT INTO challenges (school_id, type, title, description, skill, badge_label, badge_icon, sort_order, created_by)
SELECT :school_id, 'core', v.title, v.description, v.skill, v.badge_label, v.badge_icon, v.sort_order, :created_by
FROM (VALUES
  ('First Blend',
   'Mix smoothly from one track into another so the transition sounds like a single continuous groove.',
   'mixing', 'First Blend', '🎧', 10),
  ('Beatmatcher',
   'Beatmatch two tracks by ear and hold them locked for at least 16 bars — no sync button.',
   'beatmatching', 'Beatmatcher', '🥁', 20),
  ('Clean Cut',
   'Drop a crisp, on-beat cut between two tracks with no train-wreck and no dead air.',
   'transitions', 'Clean Cut', '✂️', 30),
  ('Genre Bender',
   'Blend two tracks from clearly different genres or tempos into one coherent transition.',
   'mixing', 'Genre Bender', '🎚️', 40),
  ('Loop Master',
   'Use a loop or beat-roll musically to extend a phrase and set up the next drop.',
   'performance', 'Loop Master', '🔁', 50),
  ('Crowd Reader',
   'Read the room and adapt your next selection live to keep the energy moving.',
   'selection', 'Crowd Reader', '👀', 60),
  ('Headliner',
   'Perform a complete 30-minute set with clean transitions and a deliberate energy arc.',
   'performance', 'Headliner', '🏆', 70)
) AS v(title, description, skill, badge_label, badge_icon, sort_order)
WHERE NOT EXISTS (
  SELECT 1 FROM challenges c
  WHERE c.school_id = :school_id AND c.title = v.title
);
