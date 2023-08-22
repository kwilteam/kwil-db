package atomic

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/sql"
)

// newSessionInfo creates a new sessionInfo
func newSessionInfo(dbOpener DatabaseOpener) *sessionInfo {
	return &sessionInfo{
		usedDatabases: make(map[string]Database),
		changes:       make([]*change, 0),
		dbOpener:      dbOpener,
	}
}

// sessionInfo tracks ongoing information about the session,
// such as the databases being used, changes that have been made,
// etc
type sessionInfo struct {
	// usedDatabases is a map of database ids to databases that are currently in use
	usedDatabases map[string]*usedDatabase

	// changes is a list of changes that have been made to the database in this session
	changes []*change

	// dbOpener is the database opener that is used to open databases
	dbOpener DatabaseOpener

	// savepoint
}

// usedDatabase is a database that is currently in use
type usedDatabase struct {
	db        Database
	savepoint sql.Savepoint
}

var _ changeTracker = (*sessionInfo)(nil)

// TrackChange tracks a change that was made to the database
func (s *sessionInfo) TrackChange(c *change) {
	s.changes = append(s.changes, c)
}

// GetDatabase gets the database with the given id
// if it is not cached, it will open it and add it to the cache
// it will also open up a savepoint that will be kept open for the duration of the session
func (s *sessionInfo) GetDatabase(ctx context.Context, dbid string) (Database, error) {
	db, ok := s.usedDatabases[dbid]
	if ok {
		return db.db, nil
	}

	connection, err := s.dbOpener.OpenDatabase(ctx, dbid)
	if err != nil {
		return nil, err
	}

	savepoint, err := connection.Savepoint()
	if err != nil {
		return nil, err
	}

	s.usedDatabases[dbid] = &usedDatabase{
		db:        connection,
		savepoint: savepoint,
	}

	return connection, nil
}
