package events

const (
	SchemaName      = `kwild_events` // exported because oracles depends on it. this is an issue with how oracles are unit tested
	sqlCreateSchema = `CREATE SCHEMA IF NOT EXISTS ` + SchemaName
	kvTableName     = SchemaName + `.kv`

	// eventsTable is the SQL table definition for the events table.
	// All the events in this table exist in one of the below states.
	// "received" and "broadcasted" fields are used to track the state of the event:
	// Newly added and not yet broadcasted:      [broadcasted = false and received = false]
	// Broadcasted but not included in a block:  [broadcasted = true
	// Included in a block:                      [received = true]
	eventsTable = `CREATE TABLE IF NOT EXISTS ` + SchemaName + `.events (
		id BYTEA PRIMARY KEY, -- uuid
		data BYTEA NOT NULL,
		event_type TEXT NOT NULL,
		received BOOLEAN NOT NULL DEFAULT FALSE, -- received is set to true if the network has received the vote for this event
		broadcasted BOOLEAN NOT NULL DEFAULT FALSE -- broadcasted is set to true if the event has been broadcasted by the validator. It may or may not have been received by the network
	);`
	dropEventsTable = `DROP TABLE IF EXISTS ` + SchemaName + `.events;`

	insertEventIdempotent  = `INSERT INTO ` + SchemaName + `.events (id, data, event_type) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;`
	deleteEvent            = `DELETE FROM ` + SchemaName + `.events WHERE id = $1;`
	getEvents              = `SELECT data, event_type FROM ` + SchemaName + `.events;`
	getUnbroadcastedEvents = `SELECT data, event_type FROM ` + SchemaName + `.events WHERE NOT received AND NOT broadcasted;`
	markReceived           = `UPDATE ` + SchemaName + `.events SET received = TRUE WHERE id = $1;`

	// mark list of events as broadcasted.
	markBroadcasted = `UPDATE ` + SchemaName + `.events SET broadcasted = TRUE WHERE id =ANY($1);`
	// mark list of events as not broadcasted and ready to broadcast.
	markRebroadcast = `UPDATE ` + SchemaName + `.events SET broadcasted = FALSE WHERE id =ANY($1);`

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
