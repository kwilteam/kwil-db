package events

const (
	schemaName      = `kwild_events`
	sqlCreateSchema = `CREATE SCHEMA IF NOT EXISTS ` + schemaName
	kvTableName     = schemaName + `.kv`

	eventsTable = `CREATE TABLE IF NOT EXISTS ` + schemaName + `.events (
		id BYTEA PRIMARY KEY, -- uuid
		data BYTEA NOT NULL,
		event_type TEXT NOT NULL,
		received BOOLEAN NOT NULL DEFAULT FALSE
	);`
	dropEventsTable = `DROP TABLE IF EXISTS ` + schemaName + `.events;`

	insertEventIdempotent  = `INSERT INTO ` + schemaName + `.events (id, data, event_type) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;`
	deleteEvent            = `DELETE FROM ` + schemaName + `.events WHERE id = $1;`
	getEvents              = `SELECT data, event_type FROM ` + schemaName + `.events;`
	getUnbroadcastedEvents = `SELECT data, event_type FROM ` + schemaName + `.events WHERE NOT received;`
	markReceived           = `UPDATE ` + schemaName + `.events SET received = TRUE WHERE id = $1;`
)
