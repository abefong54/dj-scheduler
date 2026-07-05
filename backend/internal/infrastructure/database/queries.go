package database

const (
	// DJ queries. Scoped to the caller's school (EL-104): a DJ belongs to the
	// school, shared by its members. EL-019 adds certifications + is_student; the
	// list supports two optional filters via params: $2 certified_for
	// (case-insensitive genre, '' = no filter) and $3 ready_only (true = only DJs
	// with ≥1 certification).
	queryDJList = `
		SELECT id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, COALESCE(email, ''), created_at::text
		FROM djs
		WHERE school_id = $1
		  AND ($2 = '' OR EXISTS (SELECT 1 FROM unnest(certifications) c WHERE LOWER(c) = LOWER($2)))
		  AND (NOT $3 OR cardinality(certifications) > 0)
		ORDER BY name`

	// DJs are school-owned (EL-104): organizer_id is legacy/nullable, so the
	// insert sets school_id only.
	queryDJInsert = `
		INSERT INTO djs (name, genre_tags, school_id) VALUES ($1, $2, $3)
		RETURNING id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, COALESCE(email, ''), created_at::text`

	queryDJGet = `
		SELECT id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, COALESCE(email, ''), created_at::text
		FROM djs WHERE id = $1 AND school_id = $2`

	// EL: the roster email (the student-login "invite") is set alongside the other
	// teacher-editable fields. NULLIF keeps an intentionally-blank email NULL so the
	// partial unique index on lower(email) doesn't collide across empty rows.
	queryDJUpdate = `
		UPDATE djs SET name = $1, genre_tags = $2, certifications = $3, is_student = $4, email = NULLIF($7, '')
		WHERE id = $5 AND school_id = $6
		RETURNING id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, COALESCE(email, ''), created_at::text`

	queryDJDelete = `DELETE FROM djs WHERE id = $1 AND school_id = $2`

	// DJ portal token queries (US-009). Only the token hash is stored.
	queryDJSetPortalToken = `
		UPDATE djs SET portal_token_hash = $1, portal_token_expires_at = $2
		WHERE id = $3 AND school_id = $4`

	queryDJGetByPortalToken = `
		SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text
		FROM djs
		WHERE portal_token_hash = $1 AND portal_token_expires_at > now()`

	// Student Google login queries (migration 014). Identity is resolved before a
	// school is known, so the two finds are intentionally NOT school-scoped — the
	// student's school comes from the matched row (school_id). email/google_id are
	// PII and are never selected here; only the columns the session needs.
	queryDJFindByGoogleID = `
		SELECT id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, school_id, created_at::text
		FROM djs WHERE google_id = $1`

	// Case-insensitive email match for the first-login auto-link.
	queryDJFindByEmail = `
		SELECT id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, school_id, created_at::text
		FROM djs WHERE lower(email) = lower($1)`

	// SetGoogleID binds the Google identity on first login. Not school-scoped:
	// djID is the row just resolved by FindByEmail.
	queryDJSetGoogleID = `UPDATE djs SET google_id = $1 WHERE id = $2`

	// SetEmail is the teacher provisioning path (the "invite"): school-scoped so a
	// teacher can only set the email of their own school's student.
	queryDJSetEmail = `UPDATE djs SET email = $1 WHERE id = $2 AND school_id = $3`

	queryDJPortalSlots = `
		SELECT sl.id, sl.event_id, e.name, st.name,
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'),
		       COALESCE(sl.notes,''),
		       sl.dj_confirmation
		FROM slots sl
		JOIN events e ON e.id = sl.event_id
		JOIN stages st ON st.id = sl.stage_id
		WHERE sl.dj_id = $1
		ORDER BY sl.slot_date, sl.start_time`

	// Event queries. Scoped to the caller's school (EL-104); organizer_id is kept
	// as a created_by audit column set on insert.
	queryEventList = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		       (line_notify_token_enc IS NOT NULL)
		FROM events WHERE school_id = $1 ORDER BY start_date DESC`

	queryEventGet = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		       (line_notify_token_enc IS NOT NULL)
		FROM events WHERE id = $1 AND school_id = $2`

	// queryEventGetPublic is intentionally NOT scoped: it backs the public,
	// shareable schedule endpoint (GET /api/events/:id/public).
	queryEventGetPublic = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE id = $1`

	// $6 organizer_id (created_by audit), $7 school_id (tenant scope).
	queryEventInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres, organizer_id, school_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		          (line_notify_token_enc IS NOT NULL)`

	queryEventUpdate = `
		UPDATE events
		SET name = $1, venue_name = $2, start_date = $3, end_date = $4, genres = $5
		WHERE id = $6 AND school_id = $7
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		          (line_notify_token_enc IS NOT NULL)`

	queryEventDelete = `DELETE FROM events WHERE id = $1 AND school_id = $2`

	// querySetLineToken stores (or clears, when $1 is NULL) the encrypted LINE
	// Notify token for the school's event, returning whether it is now enabled.
	// No row → the event isn't the school's (US-006).
	querySetLineToken = `
		UPDATE events SET line_notify_token_enc = $1
		WHERE id = $2 AND school_id = $3
		RETURNING (line_notify_token_enc IS NOT NULL)`

	// queryEventOwned reports whether an event exists AND belongs to the school.
	// Used to disambiguate an empty stage/slot list (event owned but empty → 200)
	// from a forbidden one (event not the school's → 404). See EL-036/EL-104.
	queryEventOwned = `SELECT EXISTS (SELECT 1 FROM events WHERE id = $1 AND school_id = $2)`

	// Clone (US-008): copy an event's name/venue/genres as a template. Dates are
	// reset to today on insert (the organizer sets the real dates afterwards), so
	// the source dates are not fetched. Fetch is school-scoped; insert sets
	// created_by ($4) + school_id ($5).
	queryEventCloneFetch = `
		SELECT name, venue_name, COALESCE(genres, '{}')
		FROM events WHERE id = $1 AND school_id = $2`

	queryEventCloneInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres, organizer_id, school_id)
		VALUES ('Copy of '||$1, $2, CURRENT_DATE, CURRENT_DATE, $3, $4, $5)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		          (line_notify_token_enc IS NOT NULL)`

	// Stage queries. Every stage is reachable only through its parent event, so
	// each query joins events and filters on the school (EL-104): a stage is
	// invisible (404) to anyone but its event's school, even though event UUIDs
	// are public (handed out via the shareable schedule link). See EL-036.
	queryStageList = `
		SELECT st.id, st.event_id, st.name, st.color, st.display_order
		FROM stages st
		JOIN events e ON e.id = st.event_id
		WHERE st.event_id = $1 AND e.school_id = $2
		ORDER BY st.display_order, st.name`

	queryStageInsert = `
		INSERT INTO stages (event_id, name, color)
		SELECT $1, $2, $3
		WHERE EXISTS (SELECT 1 FROM events WHERE id = $1 AND school_id = $4)
		RETURNING id, event_id, name, color, display_order`

	queryStageDelete = `
		DELETE FROM stages st
		USING events e
		WHERE st.id = $1 AND st.event_id = $2
		  AND e.id = st.event_id AND e.school_id = $3`

	queryStageGet = `
		SELECT st.id, st.event_id, st.name, st.color, st.display_order
		FROM stages st
		JOIN events e ON e.id = st.event_id
		WHERE st.id = $1 AND st.event_id = $2 AND e.school_id = $3`

	queryStageUpdate = `
		UPDATE stages st SET name = $1, color = $2
		FROM events e
		WHERE st.id = $3 AND st.event_id = $4
		  AND e.id = st.event_id AND e.school_id = $5
		RETURNING st.id, st.event_id, st.name, st.color, st.display_order`

	// queryStagePublicList is intentionally NOT scoped: it backs the public,
	// shareable schedule endpoint (GET /api/events/:id/public).
	queryStagePublicList = `
		SELECT id, event_id, name, color, display_order
		FROM stages WHERE event_id = $1 ORDER BY display_order, name`

	queryStageListForClone   = `SELECT name, color, display_order FROM stages WHERE event_id = $1 ORDER BY display_order`
	queryStageInsertForClone = `INSERT INTO stages (event_id, name, color, display_order) VALUES ($1,$2,$3,$4)`

	// Slot queries. Like stages, slots are school-scoped through their parent
	// event (EL-104): every read/write joins events and filters on school_id.
	querySlotList = `
		SELECT sl.id, sl.event_id, sl.stage_id, st.name,
		       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,''),
		       sl.dj_confirmation
		FROM slots sl
		JOIN events e ON e.id = sl.event_id
		JOIN stages st ON st.id = sl.stage_id
		LEFT JOIN djs d ON d.id = sl.dj_id
		WHERE sl.event_id = $1 AND e.school_id = $2
		ORDER BY sl.slot_date, sl.start_time`

	querySlotInsert = `
		INSERT INTO slots (event_id, stage_id, dj_id, genre, slot_date, start_time, end_time, notes)
		SELECT $1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,$8
		WHERE EXISTS (SELECT 1 FROM events WHERE id = $1 AND school_id = $9)
		RETURNING id`

	querySlotUpdate = `
		WITH updated AS (
			UPDATE slots sl
			SET stage_id=$1, dj_id=NULLIF($2,'')::uuid, genre=$3,
			    slot_date=$4, start_time=$5, end_time=$6, notes=$7
			FROM events e
			WHERE sl.id=$8 AND sl.event_id=$9
			  AND e.id = sl.event_id AND e.school_id=$10
			RETURNING sl.id, sl.event_id, sl.stage_id, sl.dj_id, sl.genre,
			          sl.slot_date, sl.start_time, sl.end_time, sl.notes, sl.dj_confirmation
		)
		SELECT u.id, u.event_id, u.stage_id, st.name,
		       COALESCE(u.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(u.genre,''),
		       u.slot_date::text, to_char(u.start_time,'HH24:MI'), to_char(u.end_time,'HH24:MI'), COALESCE(u.notes,''),
		       u.dj_confirmation
		FROM updated u
		JOIN stages st ON st.id = u.stage_id
		LEFT JOIN djs d ON d.id = u.dj_id`

	querySlotGet = `
		SELECT sl.id, sl.event_id, sl.stage_id, st.name,
		       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,''),
		       sl.dj_confirmation
		FROM slots sl
		JOIN events e ON e.id = sl.event_id
		JOIN stages st ON st.id = sl.stage_id
		LEFT JOIN djs d ON d.id = sl.dj_id
		WHERE sl.id = $1 AND sl.event_id = $2 AND e.school_id = $3`

	// querySlotPublicList is intentionally NOT scoped: it backs the public,
	// shareable schedule endpoint (GET /api/events/:id/public).
	querySlotPublicList = `
		SELECT sl.id, sl.event_id, sl.stage_id, st.name,
		       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,''),
		       sl.dj_confirmation
		FROM slots sl
		JOIN stages st ON st.id = sl.stage_id
		LEFT JOIN djs d ON d.id = sl.dj_id
		WHERE sl.event_id = $1
		ORDER BY sl.slot_date, sl.start_time`

	// querySlotPublicGet looks up a single slot purely by its id, without any
	// event/school scoping. It backs the public per-DJ share card (EL-049): the
	// slot id is the only credential. Callers must gate access some other way.
	querySlotPublicGet = `
		SELECT sl.id, sl.event_id, sl.stage_id, st.name,
		       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,''),
		       sl.dj_confirmation
		FROM slots sl
		JOIN stages st ON st.id = sl.stage_id
		LEFT JOIN djs d ON d.id = sl.dj_id
		WHERE sl.id = $1`

	// Set a DJ's confirmation on a slot the token's DJ actually owns (US-011).
	querySlotSetDJConfirmation = `
		UPDATE slots SET dj_confirmation = $1
		WHERE id = $2 AND dj_id = $3`

	querySlotDelete = `
		DELETE FROM slots sl
		USING events e
		WHERE sl.id = $1 AND sl.event_id = $2
		  AND e.id = sl.event_id AND e.school_id = $3`

	// Lead queries (EL-084). Public route — no scoping.
	queryLeadInsert = `
		INSERT INTO leads (name, organization, email, message, source)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, organization, email, message, source, created_at::text`

	// Organizer queries. google_id / line_id are nullable (a user has at least
	// one — see migration 014), so reads COALESCE them to '' for the string model.
	queryOrganizerFindByGoogleID = `
		SELECT id, email, name, COALESCE(google_id, ''), COALESCE(line_id, ''), created_at::text
		FROM organizers WHERE google_id = $1`

	queryOrganizerFindByLineID = `
		SELECT id, email, name, COALESCE(google_id, ''), COALESCE(line_id, ''), created_at::text
		FROM organizers WHERE line_id = $1`

	queryOrganizerFindByEmail = `
		SELECT id, email, name, COALESCE(google_id, ''), COALESCE(line_id, ''), created_at::text
		FROM organizers WHERE email = $1`

	queryOrganizerFindByID = `
		SELECT id, email, name, COALESCE(google_id, ''), COALESCE(line_id, ''), created_at::text
		FROM organizers WHERE id = $1`

	queryOrganizerInsert = `
		INSERT INTO organizers (email, name, google_id)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, COALESCE(google_id, ''), COALESCE(line_id, ''), created_at::text`

	queryOrganizerInsertLine = `
		INSERT INTO organizers (email, name, line_id)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, COALESCE(google_id, ''), COALESCE(line_id, ''), created_at::text`

	queryOrganizerLinkLineID = `
		UPDATE organizers SET line_id = $2 WHERE id = $1`

	// School + membership queries (EL-103, multi-tenant schools). These use
	// organizer_id as the MEMBERSHIP FK (which person), not as event/DJ scope.
	querySchoolInsert = `
		INSERT INTO schools (name, city, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, city, created_by, created_at::text`

	querySchoolGet = `
		SELECT id, name, city, created_by, created_at::text
		FROM schools WHERE id = $1`

	queryMembershipInsert = `
		INSERT INTO memberships (school_id, organizer_id, email, role, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, school_id, organizer_id, email, role, status, created_at::text`

	queryMembershipActiveForOrganizer = `
		SELECT id, school_id, organizer_id, email, role, status, created_at::text
		FROM memberships
		WHERE organizer_id = $1 AND status = 'active'
		ORDER BY created_at LIMIT 1`

	queryMembershipListBySchool = `
		SELECT id, school_id, organizer_id, email, role, status, created_at::text
		FROM memberships WHERE school_id = $1 ORDER BY created_at`

	queryMembershipLinkPendingByEmail = `
		UPDATE memberships SET organizer_id = $1, status = 'active'
		WHERE email = $2 AND organizer_id IS NULL AND status = 'pending'`

	// Performance aggregation (EL-043). Duration is not stored, so it's computed
	// from start/end TIME: a set whose end is <= start crosses midnight, so add a
	// full day to count its real length (matches usecase/slot conflict math).
	// Every query scopes through the school (EL-104).

	// queryDJPerformance: one DJ's reps across the school's events. The LEFT JOINs
	// keep the row even with zero slots (owned-but-never-played → reps 0), while a
	// DJ that isn't the school's yields no row → ErrNotFound.
	queryDJPerformance = `
		SELECT d.id, d.name,
		       COUNT(*) FILTER (WHERE sl.id IS NOT NULL AND e.id IS NOT NULL) AS reps,
		       COALESCE(SUM(
		         EXTRACT(EPOCH FROM (CASE WHEN sl.end_time <= sl.start_time
		                                  THEN (sl.end_time - sl.start_time) + INTERVAL '24 hours'
		                                  ELSE (sl.end_time - sl.start_time) END)) / 60
		       ) FILTER (WHERE e.id IS NOT NULL), 0)::int AS total_minutes,
		       COALESCE((MAX(sl.slot_date) FILTER (WHERE e.id IS NOT NULL))::text, '') AS last_played
		FROM djs d
		LEFT JOIN slots sl ON sl.dj_id = d.id
		LEFT JOIN events e ON e.id = sl.event_id AND e.school_id = $2
		WHERE d.id = $1 AND d.school_id = $2
		GROUP BY d.id, d.name`

	// queryDJPerformanceByGenre: that DJ's reps split by slot genre (empty genre
	// buckets together as "").
	queryDJPerformanceByGenre = `
		SELECT COALESCE(sl.genre, '') AS genre,
		       COUNT(*) AS reps,
		       COALESCE(SUM(
		         EXTRACT(EPOCH FROM (CASE WHEN sl.end_time <= sl.start_time
		                                  THEN (sl.end_time - sl.start_time) + INTERVAL '24 hours'
		                                  ELSE (sl.end_time - sl.start_time) END)) / 60
		       ), 0)::int AS total_minutes
		FROM slots sl
		JOIN events e ON e.id = sl.event_id
		WHERE sl.dj_id = $1 AND e.school_id = $2
		GROUP BY COALESCE(sl.genre, '')
		ORDER BY reps DESC, genre`

	// queryRosterSummary: every active student (is_student = true) with reps in
	// the window, including zero-rep students. The event/date filters live in the
	// LEFT JOIN ON clauses so non-matching students still appear with reps 0.
	// $1 school, $2 event_id (''=all), $3 from (''=none), $4 to (''=none).
	queryRosterSummary = `
		SELECT d.id, d.name, d.is_student,
		       COUNT(*) FILTER (WHERE sl.id IS NOT NULL AND e.id IS NOT NULL) AS reps,
		       COALESCE(SUM(
		         EXTRACT(EPOCH FROM (CASE WHEN sl.end_time <= sl.start_time
		                                  THEN (sl.end_time - sl.start_time) + INTERVAL '24 hours'
		                                  ELSE (sl.end_time - sl.start_time) END)) / 60
		       ) FILTER (WHERE e.id IS NOT NULL), 0)::int AS total_minutes,
		       COALESCE((MAX(sl.slot_date) FILTER (WHERE e.id IS NOT NULL))::text, '') AS last_played
		FROM djs d
		LEFT JOIN slots sl ON sl.dj_id = d.id
		       AND ($3 = '' OR sl.slot_date >= NULLIF($3, '')::date)
		       AND ($4 = '' OR sl.slot_date <= NULLIF($4, '')::date)
		LEFT JOIN events e ON e.id = sl.event_id
		       AND e.school_id = $1
		       AND ($2 = '' OR e.id = NULLIF($2, '')::uuid)
		WHERE d.school_id = $1 AND d.is_student = true
		GROUP BY d.id, d.name, d.is_student
		ORDER BY reps ASC, d.name`

	// Challenge queries (migration 015). Challenges are school-owned, so every
	// read/write filters on school_id — a challenge that isn't the caller's school's
	// simply doesn't match and reads as not-found. The nullable columns (skill,
	// assigned_dj_id, badge_label, badge_icon) are COALESCE'd to '' on read and
	// NULLIF('')'d on write, mirroring the slot queries.
	challengeCols = `id, school_id, type, title, description,
		COALESCE(skill,''), COALESCE(assigned_dj_id::text,''),
		COALESCE(badge_label,''), COALESCE(badge_icon,''),
		sort_order, created_by, archived, created_at::text`

	queryChallengeList = `
		SELECT ` + challengeCols + `
		FROM challenges
		WHERE school_id = $1 AND NOT archived
		ORDER BY sort_order, created_at`

	queryChallengeGet = `
		SELECT ` + challengeCols + `
		FROM challenges
		WHERE id = $1 AND school_id = $2`

	queryChallengeInsert = `
		INSERT INTO challenges
			(school_id, type, title, description, skill, assigned_dj_id, badge_label, badge_icon, sort_order, created_by)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,'')::uuid,NULLIF($7,''),NULLIF($8,''),$9,$10)
		RETURNING id, created_at::text`

	queryChallengeUpdate = `
		UPDATE challenges
		SET type=$1, title=$2, description=$3, skill=NULLIF($4,''),
		    assigned_dj_id=NULLIF($5,'')::uuid, badge_label=NULLIF($6,''),
		    badge_icon=NULLIF($7,''), sort_order=$8
		WHERE id=$9 AND school_id=$10
		RETURNING ` + challengeCols

	queryChallengeArchive = `
		UPDATE challenges SET archived = TRUE
		WHERE id = $1 AND school_id = $2`

	// Core (badge-bearing, school-wide) challenges for "Browse Core".
	queryChallengeListCore = `
		SELECT ` + challengeCols + `
		FROM challenges
		WHERE school_id = $1 AND type = 'core' AND NOT archived
		ORDER BY sort_order, created_at`

	// A student's own open custom challenges: the ones assigned to them, scoped to
	// their school. $1 dj_id, $2 school_id. The usecase merges these with the core
	// set to form the student's full challenge list.
	queryChallengeListForStudent = `
		SELECT ` + challengeCols + `
		FROM challenges
		WHERE assigned_dj_id = $1 AND school_id = $2
		  AND type = 'custom' AND NOT archived
		ORDER BY sort_order, created_at`
)

// submissionCols is the SELECT/RETURNING column list for submissions, with
// nullable columns COALESCE'd to ” and timestamps/uuids cast to text so the
// model can hold plain strings (mirrors challengeCols).
const submissionCols = `id, school_id, challenge_id, dj_id, media_kind, media_provider, media_ref,
	COALESCE(caption,''), status, COALESCE(review_note,''),
	COALESCE(earned_badge_label,''), COALESCE(earned_badge_icon,''),
	COALESCE(reviewed_by::text,''), COALESCE(reviewed_at::text,''), created_at::text`

const (
	querySubmissionGet = `
		SELECT ` + submissionCols + `
		FROM submissions
		WHERE id = $1 AND school_id = $2`

	// querySubmissionUpsert writes the one active submission per (dj_id,
	// challenge_id): a re-post overwrites the media and resets the row to
	// 'posted', clearing any prior review/badge snapshot.
	querySubmissionUpsert = `
		INSERT INTO submissions
			(school_id, challenge_id, dj_id, media_kind, media_provider, media_ref, caption, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'posted')
		ON CONFLICT (dj_id, challenge_id) DO UPDATE SET
			media_kind = EXCLUDED.media_kind,
			media_provider = EXCLUDED.media_provider,
			media_ref = EXCLUDED.media_ref,
			caption = EXCLUDED.caption,
			status = 'posted',
			review_note = NULL,
			earned_badge_label = NULL,
			earned_badge_icon = NULL,
			reviewed_by = NULL,
			reviewed_at = NULL
		RETURNING ` + submissionCols

	// querySubmissionReview records a teacher pass/needs_work. On a pass, the
	// badge is snapshotted from the parent challenge IN SQL, but only when that
	// challenge is core — so the earned badge is frozen onto the row and survives
	// later edits/archival of the challenge.
	querySubmissionReview = `
		UPDATE submissions SET
			status = $3,
			review_note = NULLIF($4,''),
			reviewed_at = now(),
			earned_badge_label = CASE WHEN $3 = 'passed'
				THEN (SELECT badge_label FROM challenges c WHERE c.id = submissions.challenge_id AND c.type = 'core')
				ELSE NULL END,
			earned_badge_icon = CASE WHEN $3 = 'passed'
				THEN (SELECT badge_icon FROM challenges c WHERE c.id = submissions.challenge_id AND c.type = 'core')
				ELSE NULL END
		WHERE id = $1 AND school_id = $2
		RETURNING ` + submissionCols

	// querySubmissionListFeed is the school feed: passed submissions only, newest
	// review first.
	querySubmissionListFeed = `
		SELECT ` + submissionCols + `
		FROM submissions
		WHERE school_id = $1 AND status = 'passed'
		ORDER BY reviewed_at DESC NULLS LAST, created_at DESC`

	querySubmissionListForStudent = `
		SELECT ` + submissionCols + `
		FROM submissions
		WHERE dj_id = $1 AND school_id = $2
		ORDER BY created_at DESC`

	// querySubmissionListForReview is the teacher review queue: posted (awaiting
	// sign-off) submissions, oldest first so the longest-waiting student is served.
	querySubmissionListForReview = `
		SELECT ` + submissionCols + `
		FROM submissions
		WHERE school_id = $1 AND status = 'posted'
		ORDER BY created_at ASC`
)
