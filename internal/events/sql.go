package events

var (
	eventsTable = `CREATE TABLE IF NOT EXISTS events (
		id BLOB PRIMARY KEY, -- uuid
		data BLOB NOT NULL,
		event_type TEXT NOT NULL
	) WITHOUT ROWID, STRICT;`

	insertEventIdempotent = `INSERT INTO events (id, data, event_type) VALUES ($id, $data, $event_type) ON CONFLICT DO NOTHING;`
	deleteEvent           = `DELETE FROM events WHERE id = $id;`
	getEvents             = `SELECT data, event_type FROM events;`
)
