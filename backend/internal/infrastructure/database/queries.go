package database

const (
	// DJ queries. EL-019 adds certifications + is_student; the list supports two
	// optional filters via params: $2 certified_for (case-insensitive genre, ''
	// = no filter) and $3 ready_only (true = only DJs with ≥1 certification).
	queryDJList = `
		SELECT id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, created_at::text
		FROM djs
		WHERE organizer_id = $1
		  AND ($2 = '' OR EXISTS (SELECT 1 FROM unnest(certifications) c WHERE LOWER(c) = LOWER($2)))
		  AND (NOT $3 OR cardinality(certifications) > 0)
		ORDER BY name`

	queryDJInsert = `
		INSERT INTO djs (name, genre_tags, organizer_id) VALUES ($1, $2, $3)
		RETURNING id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, created_at::text`

	queryDJGet = `
		SELECT id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, created_at::text
		FROM djs WHERE id = $1 AND organizer_id = $2`

	queryDJUpdate = `
		UPDATE djs SET name = $1, genre_tags = $2, certifications = $3, is_student = $4
		WHERE id = $5 AND organizer_id = $6
		RETURNING id, name, COALESCE(genre_tags, '{}'), COALESCE(certifications, '{}'), is_student, created_at::text`

	queryDJDelete = `DELETE FROM djs WHERE id = $1 AND organizer_id = $2`

	// DJ portal token queries (US-009). Only the token hash is stored.
	queryDJSetPortalToken = `
		UPDATE djs SET portal_token_hash = $1, portal_token_expires_at = $2
		WHERE id = $3 AND organizer_id = $4`

	queryDJGetByPortalToken = `
		SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text
		FROM djs
		WHERE portal_token_hash = $1 AND portal_token_expires_at > now()`

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

	// Event queries
	queryEventList = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		       (line_notify_token_enc IS NOT NULL)
		FROM events WHERE organizer_id = $1 ORDER BY start_date DESC`

	queryEventGet = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		       (line_notify_token_enc IS NOT NULL)
		FROM events WHERE id = $1 AND organizer_id = $2`

	// queryEventGetPublic is intentionally NOT scoped by organizer: it backs the
	// public, shareable schedule endpoint (GET /api/events/:id/public).
	queryEventGetPublic = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE id = $1`

	queryEventInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres, organizer_id)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		          (line_notify_token_enc IS NOT NULL)`

	queryEventUpdate = `
		UPDATE events
		SET name = $1, venue_name = $2, start_date = $3, end_date = $4, genres = $5
		WHERE id = $6 AND organizer_id = $7
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		          (line_notify_token_enc IS NOT NULL)`

	queryEventDelete = `DELETE FROM events WHERE id = $1 AND organizer_id = $2`

	// querySetLineToken stores (or clears, when $1 is NULL) the encrypted LINE
	// Notify token for the organizer's event, returning whether it is now
	// enabled. No row → the event isn't the organizer's (US-006).
	querySetLineToken = `
		UPDATE events SET line_notify_token_enc = $1
		WHERE id = $2 AND organizer_id = $3
		RETURNING (line_notify_token_enc IS NOT NULL)`

	// queryEventOwned reports whether an event exists AND belongs to the organizer.
	// Used to disambiguate an empty stage/slot list (event owned but empty → 200)
	// from a forbidden one (event not the organizer's → 404). See EL-036.
	queryEventOwned = `SELECT EXISTS (SELECT 1 FROM events WHERE id = $1 AND organizer_id = $2)`

	// Clone (US-008): copy an event's name/venue/genres as a template. Dates are
	// reset to today on insert (the organizer sets the real dates afterwards), so
	// the source dates are not fetched.
	queryEventCloneFetch = `
		SELECT name, venue_name, COALESCE(genres, '{}')
		FROM events WHERE id = $1 AND organizer_id = $2`

	queryEventCloneInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres, organizer_id)
		VALUES ('Copy of '||$1, $2, CURRENT_DATE, CURRENT_DATE, $3, $4)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}'),
		          (line_notify_token_enc IS NOT NULL)`

	// Stage queries. Every stage is reachable only through its parent event, so
	// each query joins events and filters on organizer_id: a stage is invisible
	// (404) to anyone but its event's owner, even though event UUIDs are public
	// (handed out via the shareable schedule link). See EL-036.
	queryStageList = `
		SELECT st.id, st.event_id, st.name, st.color, st.display_order
		FROM stages st
		JOIN events e ON e.id = st.event_id
		WHERE st.event_id = $1 AND e.organizer_id = $2
		ORDER BY st.display_order, st.name`

	queryStageInsert = `
		INSERT INTO stages (event_id, name, color)
		SELECT $1, $2, $3
		WHERE EXISTS (SELECT 1 FROM events WHERE id = $1 AND organizer_id = $4)
		RETURNING id, event_id, name, color, display_order`

	queryStageDelete = `
		DELETE FROM stages st
		USING events e
		WHERE st.id = $1 AND st.event_id = $2
		  AND e.id = st.event_id AND e.organizer_id = $3`

	queryStageGet = `
		SELECT st.id, st.event_id, st.name, st.color, st.display_order
		FROM stages st
		JOIN events e ON e.id = st.event_id
		WHERE st.id = $1 AND st.event_id = $2 AND e.organizer_id = $3`

	queryStageUpdate = `
		UPDATE stages st SET name = $1, color = $2
		FROM events e
		WHERE st.id = $3 AND st.event_id = $4
		  AND e.id = st.event_id AND e.organizer_id = $5
		RETURNING st.id, st.event_id, st.name, st.color, st.display_order`

	// queryStagePublicList is intentionally NOT organizer-scoped: it backs the
	// public, shareable schedule endpoint (GET /api/events/:id/public), the same
	// way queryEventGetPublic does for the event itself.
	queryStagePublicList = `
		SELECT id, event_id, name, color, display_order
		FROM stages WHERE event_id = $1 ORDER BY display_order, name`

	queryStageListForClone   = `SELECT name, color, display_order FROM stages WHERE event_id = $1 ORDER BY display_order`
	queryStageInsertForClone = `INSERT INTO stages (event_id, name, color, display_order) VALUES ($1,$2,$3,$4)`

	// Slot queries. Like stages, slots are organizer-scoped through their parent
	// event (EL-036): every read/write joins events and filters on organizer_id.
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
		WHERE sl.event_id = $1 AND e.organizer_id = $2
		ORDER BY sl.slot_date, sl.start_time`

	querySlotInsert = `
		INSERT INTO slots (event_id, stage_id, dj_id, genre, slot_date, start_time, end_time, notes)
		SELECT $1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,$8
		WHERE EXISTS (SELECT 1 FROM events WHERE id = $1 AND organizer_id = $9)
		RETURNING id`

	querySlotUpdate = `
		WITH updated AS (
			UPDATE slots sl
			SET stage_id=$1, dj_id=NULLIF($2,'')::uuid, genre=$3,
			    slot_date=$4, start_time=$5, end_time=$6, notes=$7
			FROM events e
			WHERE sl.id=$8 AND sl.event_id=$9
			  AND e.id = sl.event_id AND e.organizer_id=$10
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
		WHERE sl.id = $1 AND sl.event_id = $2 AND e.organizer_id = $3`

	// querySlotPublicList is intentionally NOT organizer-scoped: it backs the
	// public, shareable schedule endpoint (GET /api/events/:id/public).
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
	// organizer/event scoping. It backs the public per-DJ share card (EL-049):
	// the slot id is the only credential. Callers must gate access some other way.
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
		  AND e.id = sl.event_id AND e.organizer_id = $3`

	// Organizer queries
	queryOrganizerFindByGoogleID = `
		SELECT id, email, name, google_id, created_at::text
		FROM organizers WHERE google_id = $1`

	queryOrganizerInsert = `
		INSERT INTO organizers (email, name, google_id)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, google_id, created_at::text`

	// Performance aggregation (EL-043). Duration is not stored, so it's computed
	// from start/end TIME: a set whose end is <= start crosses midnight, so add a
	// full day to count its real length (matches usecase/slot conflict math).
	// Every query scopes through events.organizer_id (EL-036).

	// queryDJPerformance: one DJ's reps across the organizer's events. The LEFT
	// JOINs keep the row even with zero slots (owned-but-never-played → reps 0),
	// while a DJ that isn't the organizer's yields no row → ErrNotFound.
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
		LEFT JOIN events e ON e.id = sl.event_id AND e.organizer_id = $2
		WHERE d.id = $1 AND d.organizer_id = $2
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
		WHERE sl.dj_id = $1 AND e.organizer_id = $2
		GROUP BY COALESCE(sl.genre, '')
		ORDER BY reps DESC, genre`

	// queryRosterSummary: every active student (is_student = true) with reps in
	// the window, including zero-rep students. The event/date filters live in the
	// LEFT JOIN ON clauses so non-matching students still appear with reps 0.
	// $1 organizer, $2 event_id (''=all), $3 from (''=none), $4 to (''=none).
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
		       AND e.organizer_id = $1
		       AND ($2 = '' OR e.id = NULLIF($2, '')::uuid)
		WHERE d.organizer_id = $1 AND d.is_student = true
		GROUP BY d.id, d.name, d.is_student
		ORDER BY reps ASC, d.name`
)
