package registry

import (
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// Session is a session.
// It exists for the lifetime of a commit.
type session struct {
	// Open is a map of open connections.
	// They will be closed at the end of the session.
	Open map[string]*openDB

	// IdempotencyKey is the idempotency key for the session.
	IdempotencyKey []byte

	// If Recovery is true, the database is in recovery mode.
	Recovery bool

	// Committed is the set of committed databases
	// It maps the dbid to the app hash
	Committed map[string][]byte
}

func newSession(idempotencyKey []byte, recovery bool) *session {
	return &session{
		Open:           make(map[string]*openDB),
		IdempotencyKey: idempotencyKey,
		Recovery:       recovery,
		Committed:      make(map[string][]byte),
	}
}

// openDB is a database that has been used in a session.
type openDB struct {
	// Pool is the writer connection to the database.
	Pool Pool

	// Savepoint is the savepoint for the connection.
	Savepoint sql.Savepoint

	// Session is the session for the connection.
	Session sql.Session

	// Status tracks whether the database was created or deleted.
	Status dbStatus
}

type dbStatus uint8

const (
	// dbStatusExists means a database existed before the session, and should exist after the session.
	dbStatusExists dbStatus = iota
	// dbStatusNew means a database was created in the session.
	dbStatusNew
	// dbStatusDeleted means a database was deleted in the session.
	dbStatusDeleted
)
