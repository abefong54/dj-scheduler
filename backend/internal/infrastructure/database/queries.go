package database

const (
	// DJ queries
	queryDJList = `SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text FROM djs WHERE organizer_id = $1 ORDER BY name`

	queryDJInsert = `
		INSERT INTO djs (name, genre_tags, organizer_id) VALUES ($1, $2, $3)
		RETURNING id, name, COALESCE(genre_tags, '{}'), created_at::text`

	queryDJGet = `
		SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text
		FROM djs WHERE id = $1 AND organizer_id = $2`

	queryDJUpdate = `
		UPDATE djs SET name = $1, genre_tags = $2
		WHERE id = $3 AND organizer_id = $4
		RETURNING id, name, COALESCE(genre_tags, '{}'), created_at::text`

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
		SELECT sl.event_id, e.name, st.name,
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'),
		       COALESCE(sl.notes,'')
		FROM slots sl
		JOIN events e ON e.id = sl.event_id
		JOIN stages st ON st.id = sl.stage_id
		WHERE sl.dj_id = $1
		ORDER BY sl.slot_date, sl.start_time`

	// Event queries
	queryEventList = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE organizer_id = $1 ORDER BY start_date DESC`

	queryEventGet = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE id = $1 AND organizer_id = $2`

	// queryEventGetPublic is intentionally NOT scoped by organizer: it backs the
	// public, shareable schedule endpoint (GET /api/events/:id/public).
	queryEventGetPublic = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE id = $1`

	queryEventInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres, organizer_id)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')`

	queryEventUpdate = `
		UPDATE events
		SET name = $1, venue_name = $2, start_date = $3, end_date = $4, genres = $5
		WHERE id = $6 AND organizer_id = $7
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')`

	queryEventDelete = `DELETE FROM events WHERE id = $1 AND organizer_id = $2`

	// Clone (US-008): copy an event's name/venue/genres as a template. Dates are
	// reset to today on insert (the organizer sets the real dates afterwards), so
	// the source dates are not fetched.
	queryEventCloneFetch = `
		SELECT name, venue_name, COALESCE(genres, '{}')
		FROM events WHERE id = $1 AND organizer_id = $2`

	queryEventCloneInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres, organizer_id)
		VALUES ('Copy of '||$1, $2, CURRENT_DATE, CURRENT_DATE, $3, $4)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')`

	// Stage queries
	queryStageList = `
		SELECT id, event_id, name, color, display_order
		FROM stages WHERE event_id = $1 ORDER BY display_order, name`

	queryStageInsert = `
		INSERT INTO stages (event_id, name, color)
		VALUES ($1,$2,$3)
		RETURNING id, event_id, name, color, display_order`

	queryStageDelete = `DELETE FROM stages WHERE id = $1 AND event_id = $2`

	queryStageGet = `
		SELECT id, event_id, name, color, display_order
		FROM stages WHERE id = $1 AND event_id = $2`

	queryStageUpdate = `
		UPDATE stages SET name = $1, color = $2
		WHERE id = $3 AND event_id = $4
		RETURNING id, event_id, name, color, display_order`

	queryStageListForClone   = `SELECT name, color, display_order FROM stages WHERE event_id = $1 ORDER BY display_order`
	queryStageInsertForClone = `INSERT INTO stages (event_id, name, color, display_order) VALUES ($1,$2,$3,$4)`

	// Slot queries
	querySlotList = `
		SELECT sl.id, sl.event_id, sl.stage_id, st.name,
		       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,'')
		FROM slots sl
		JOIN stages st ON st.id = sl.stage_id
		LEFT JOIN djs d ON d.id = sl.dj_id
		WHERE sl.event_id = $1
		ORDER BY sl.slot_date, sl.start_time`

	querySlotInsert = `
		INSERT INTO slots (event_id, stage_id, dj_id, genre, slot_date, start_time, end_time, notes)
		VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,$8)
		RETURNING id`

	querySlotUpdate = `
		WITH updated AS (
			UPDATE slots
			SET stage_id=$1, dj_id=NULLIF($2,'')::uuid, genre=$3,
			    slot_date=$4, start_time=$5, end_time=$6, notes=$7
			WHERE id=$8 AND event_id=$9
			RETURNING *
		)
		SELECT u.id, u.event_id, u.stage_id, st.name,
		       COALESCE(u.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(u.genre,''),
		       u.slot_date::text, to_char(u.start_time,'HH24:MI'), to_char(u.end_time,'HH24:MI'), COALESCE(u.notes,'')
		FROM updated u
		JOIN stages st ON st.id = u.stage_id
		LEFT JOIN djs d ON d.id = u.dj_id`

	querySlotGet = `
		SELECT sl.id, sl.event_id, sl.stage_id, st.name,
		       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
		       COALESCE(sl.genre,''),
		       sl.slot_date::text, to_char(sl.start_time,'HH24:MI'), to_char(sl.end_time,'HH24:MI'), COALESCE(sl.notes,'')
		FROM slots sl
		JOIN stages st ON st.id = sl.stage_id
		LEFT JOIN djs d ON d.id = sl.dj_id
		WHERE sl.id = $1 AND sl.event_id = $2`

	querySlotDelete = `DELETE FROM slots WHERE id = $1 AND event_id = $2`

	// Organizer queries
	queryOrganizerFindByGoogleID = `
		SELECT id, email, name, google_id, created_at::text
		FROM organizers WHERE google_id = $1`

	queryOrganizerInsert = `
		INSERT INTO organizers (email, name, google_id)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, google_id, created_at::text`
)
