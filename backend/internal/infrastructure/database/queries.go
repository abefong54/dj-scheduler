package database

const (
	// DJ queries
	queryDJList = `SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text FROM djs ORDER BY name`

	queryDJInsert = `
		INSERT INTO djs (name, genre_tags) VALUES ($1, $2)
		RETURNING id, name, COALESCE(genre_tags, '{}'), created_at::text`

	queryDJGet = `
		SELECT id, name, COALESCE(genre_tags, '{}'), created_at::text
		FROM djs WHERE id = $1`

	queryDJUpdate = `
		UPDATE djs SET name = $1, genre_tags = $2
		WHERE id = $3
		RETURNING id, name, COALESCE(genre_tags, '{}'), created_at::text`

	queryDJDelete = `DELETE FROM djs WHERE id = $1`

	// Event queries
	queryEventList = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events ORDER BY start_date DESC`

	queryEventGet = `
		SELECT id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE id = $1`

	queryEventInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')`

	queryEventDelete = `DELETE FROM events WHERE id = $1`

	queryEventDuplicateFetch = `
		SELECT name, venue_name, start_date::text, end_date::text, COALESCE(genres, '{}')
		FROM events WHERE id = $1`

	queryEventDuplicateInsert = `
		INSERT INTO events (name, venue_name, start_date, end_date, genres)
		VALUES ($1||' (copy)', $2, $3, $4, $5)
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

	queryStageListForDuplicate   = `SELECT name, color, display_order FROM stages WHERE event_id = $1 ORDER BY display_order`
	queryStageInsertForDuplicate = `INSERT INTO stages (event_id, name, color, display_order) VALUES ($1,$2,$3,$4)`

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

	querySlotDelete = `DELETE FROM slots WHERE id = $1 AND event_id = $2`
)
