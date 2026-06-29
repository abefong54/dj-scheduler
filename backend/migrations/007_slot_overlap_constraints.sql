-- backend/migrations/007_slot_overlap_constraints.sql
-- Make slot non-overlap a database invariant, not just an application check.
--
-- The use case checks conflicts before writing, but a check-then-write has a
-- race: two concurrent requests can both pass the check and both insert. These
-- GiST exclusion constraints let Postgres reject the overlapping write itself,
-- so exactly one of any racing pair succeeds (BUG-006).
--
-- Each slot's real interval is [slot_date + start_time, slot_date + end_time),
-- with a day added when end_time <= start_time (the set runs past midnight —
-- BUG-004). Using absolute timestamps makes the range comparison handle the
-- cross-midnight and cross-date cases for free.

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- No two slots on the same stage may overlap in real time.
ALTER TABLE slots ADD CONSTRAINT slots_stage_no_overlap
    EXCLUDE USING gist (
        stage_id WITH =,
        tsrange(
            slot_date + start_time,
            slot_date + end_time
                + (CASE WHEN end_time <= start_time THEN INTERVAL '1 day' ELSE INTERVAL '0 day' END)
        ) WITH &&
    );

-- A DJ may not be booked into two overlapping slots (any stage). Unassigned
-- slots (dj_id IS NULL) are exempt — an empty slot can't double-book anyone.
ALTER TABLE slots ADD CONSTRAINT slots_dj_no_overlap
    EXCLUDE USING gist (
        dj_id WITH =,
        tsrange(
            slot_date + start_time,
            slot_date + end_time
                + (CASE WHEN end_time <= start_time THEN INTERVAL '1 day' ELSE INTERVAL '0 day' END)
        ) WITH &&
    ) WHERE (dj_id IS NOT NULL);
