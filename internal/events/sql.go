package events

const (
	SchemaName      = `kwild_events` // exported because oracles depends on it. this is an issue with how oracles are unit tested
	sqlCreateSchema = `CREATE SCHEMA IF NOT EXISTS ` + SchemaName
	kvTableName     = SchemaName + `.kv`

	eventsTable = `CREATE TABLE IF NOT EXISTS ` + SchemaName + `.events (
		id BYTEA PRIMARY KEY, -- uuid
		data BYTEA NOT NULL,
		event_type TEXT NOT NULL,
		received BOOLEAN NOT NULL DEFAULT FALSE
	);`
	dropEventsTable = `DROP TABLE IF EXISTS ` + SchemaName + `.events;`

	insertEventIdempotent  = `INSERT INTO ` + SchemaName + `.events (id, data, event_type) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;`
	deleteEvent            = `DELETE FROM ` + SchemaName + `.events WHERE id = $1;`
	getEvents              = `SELECT data, event_type FROM ` + SchemaName + `.events;`
	getUnbroadcastedEvents = `SELECT data, event_type FROM ` + SchemaName + `.events WHERE NOT received;`
	markReceived           = `UPDATE ` + SchemaName + `.events SET received = TRUE WHERE id = $1;`

	// KV sql
	kvTableNameFull = SchemaName + "_kv"
	createKvTblStmt = `
	CREATE TABLE IF NOT EXISTS ` + kvTableNameFull + ` (
		key BYTEA PRIMARY KEY,
		value BYTEA NOT NULL
	);`
	upsertKvStmt = `
		INSERT INTO ` + kvTableNameFull + ` (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2;
	`
	selectKvStmt = `
		SELECT value
		FROM ` + kvTableNameFull + `
		WHERE key = $1;
	`
	deleteKvStmt = `
		DELETE FROM ` + kvTableNameFull + `
		WHERE key = $1;
	`
)
