package atomic

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/pkg/sql"
)

// newSessionInfo creates a new sessionInfo
func newSessionInfo(dbOpener DatabaseOpener) *sessionInfo {
	return &sessionInfo{
		usedDatabases: make(map[string]*usedDatabase),
		changes:       make([]*change, 0),
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
}

// usedDatabase is a database that is currently in use
type usedDatabase struct {
	// db is the database connection
	db Database

	// savepoint is the outermost savepoint that is currently open
	// it is opened when the database is opened, and closed when the session is closed
	savepoint sql.Savepoint
}

var _ changeTracker = (*sessionInfo)(nil)

// TrackChange tracks a change that was made to the database
func (s *sessionInfo) TrackChange(c *change) {
	s.changes = append(s.changes, c)
}

// RegisterDatabase registers a database with the session
// if the database is already registered, this method does nothing
func (s *sessionInfo) RegisterDatabase(ctx context.Context, dbid string, db Database) error {
	_, ok := s.usedDatabases[dbid]
	if ok {
		return nil
	}

	savepoint, err := db.Savepoint()
	if err != nil {
		return err
	}

	s.usedDatabases[dbid] = &usedDatabase{
		db:        db,
		savepoint: savepoint,
	}

	return nil
}

// GetDatabase gets a database that is currently registered with the session
// If no database is registered with the given dbid, this method returns false
func (s *sessionInfo) GetDatabase(dbid string) (Database, bool) {
	db, ok := s.usedDatabases[dbid]
	if !ok {
		return nil, false
	}

	return db.db, true
}

// Commit commits all of the savepoints that are currently open, and checkpoints the databases
func (s *sessionInfo) Commit() error {
	for _, db := range s.usedDatabases {
		err := db.savepoint.Commit()
		if err != nil {
			s.Rollback() // attempt to rollback all of the savepoints, but ignore the error since
			// already committed savepoints will return an error
			return err
		}
	}

	// we want to run this in a separate loop so all checkpoints are run after all commits
	for _, db := range s.usedDatabases {
		err := db.db.CheckpointWal()
		if err != nil {
			return err
		}
	}

	return nil
}

// RollbackAndReopen rolls back all of the savepoints that are currently open, and reopens new savepoints for them
func (s *sessionInfo) RollbackAndReopen() error {
	// We will try to roll all back, and return any errors at the end
	var errs []error
	for _, db := range s.usedDatabases {
		err := db.savepoint.Rollback()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if errors.Join(errs...) != nil {
		return errors.Join(errs...)
	}

	// tracks the savepoints that have been opened
	openedSPs := []string{}
	// If any fail, we should just return the error
	for dbid, db := range s.usedDatabases {
		savepoint, err := db.db.Savepoint()
		if err != nil {
			// close all of the savepoints that have been opened
			for _, db := range openedSPs {
				err2 := s.usedDatabases[db].savepoint.Rollback()
				if err2 != nil {
					err = errors.Join(err, err2)
				}
			}

			return err
		}

		openedSPs = append(openedSPs, dbid)
		db.savepoint = savepoint
	}

	return nil
}

// Rollback rolls back all of the savepoints that are currently open
func (s *sessionInfo) Rollback() error {
	// We will try to roll all back, and return any errors at the end
	var errs []error
	for _, db := range s.usedDatabases {
		err := db.savepoint.Rollback()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Clear clears all session information
func (s *sessionInfo) Clear() {
	clear(s.usedDatabases)
	clear(s.changes)
}
