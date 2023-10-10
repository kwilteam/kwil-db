package sqlsession

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sessions"
	"github.com/kwilteam/kwil-db/internal/sql"
	"go.uber.org/zap"
)

type SqlCommitable struct {
	db SqlDB

	session   sql.Session
	savepoint sql.Savepoint

	log log.Logger
}

var _ sessions.Committable = (*SqlCommitable)(nil)

// NewSqlCommitable creates a new SqlCommitable.
func NewSqlCommittable(db SqlDB, opts ...SqlCommittableOpt) *SqlCommitable {
	s := &SqlCommitable{
		db:  db,
		log: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// BeginCommit starts a new session.
func (s *SqlCommitable) BeginCommit(ctx context.Context) error {
	if s.session != nil {
		return fmt.Errorf("session already started")
	}
	if s.savepoint != nil {
		return fmt.Errorf("savepoint already active")
	}

	session, err := s.db.CreateSession()
	if err != nil {
		return err
	}
	s.session = session

	savepoint, err := s.db.Savepoint()
	if err != nil {
		s.session = nil
		return err
	}

	s.savepoint = savepoint

	return nil
}

// EndCommit ends the current session and commits the changes.
func (s *SqlCommitable) EndCommit(ctx context.Context, appender func([]byte) error) (err error) {
	if s.session == nil {
		return fmt.Errorf("session not started")
	}
	if s.savepoint == nil {
		return fmt.Errorf("savepoint not active")
	}

	defer func() {
		err = errors.Join(s.savepoint.Rollback(), s.session.Delete())
		s.savepoint = nil
		s.session = nil
		if err != nil {
			s.log.Error("failed to clean up sql committable", zap.Error(err))
		}
	}()

	changes, err := s.session.GenerateChangeset()
	if err != nil {
		return err
	}
	defer changes.Close()

	data, err := changes.Export()
	if err != nil {
		return err
	}

	return appender(data)
}

// BeginApply starts a new savepoint for applying changes.
func (s *SqlCommitable) BeginApply(ctx context.Context) error {
	if s.savepoint != nil {
		return fmt.Errorf("savepoint already active")
	}

	err := s.db.DisableForeignKey()
	if err != nil {
		s.db.EnableForeignKey()
		return err
	}

	savepoint, err := s.db.Savepoint()
	if err != nil {
		s.db.EnableForeignKey()
		return err
	}

	s.savepoint = savepoint

	return nil
}

// Apply applies a change to the database.
func (s *SqlCommitable) Apply(ctx context.Context, changes []byte) error {
	if s.savepoint == nil {
		return fmt.Errorf("savepoint not active")
	}

	return s.db.ApplyChangeset(bytes.NewReader(changes))
}

// EndApply ends the current savepoint.
func (s *SqlCommitable) EndApply(ctx context.Context) error {
	defer s.db.EnableForeignKey()
	if s.savepoint == nil {
		return fmt.Errorf("savepoint not active")
	}

	err := s.savepoint.Commit()
	if err != nil {
		return err
	}

	s.savepoint = nil

	return s.db.CheckpointWal()
}

// Cancel cancels the current session.
// it deletes the session and rolls back the savepoint.
// it will also enable foreign key constraints.
func (s *SqlCommitable) Cancel(ctx context.Context) error {
	var errs error
	if s.session != nil {
		errs = errors.Join(errs, s.session.Delete())
		s.session = nil
	}
	if s.savepoint != nil {
		errs = errors.Join(errs, s.savepoint.Rollback())
		s.savepoint = nil
	}

	// this can be called multiple times without issue
	if err := s.db.EnableForeignKey(); err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		s.log.Error("errors while cancelling session", zap.Error(errs))
	}
	return errs
}

// ID returns the ID of the current session.
func (s *SqlCommitable) ID(ctx context.Context) ([]byte, error) {
	if s.session == nil {
		return nil, fmt.Errorf("session not started")
	}
	if s.savepoint == nil {
		return nil, fmt.Errorf("savepoint not active")
	}

	changes, err := s.session.GenerateChangeset()
	if err != nil {
		return nil, err
	}
	defer changes.Close()

	return changes.ID()
}

type SqlCommittableOpt func(*SqlCommitable)

func WithLogger(logger log.Logger) SqlCommittableOpt {
	return func(s *SqlCommitable) {
		s.log = logger
	}
}

type SqlDB interface {
	ApplyChangeset(reader io.Reader) error
	CreateSession() (sql.Session, error)
	Savepoint() (sql.Savepoint, error)
	CheckpointWal() error
	EnableForeignKey() error
	DisableForeignKey() error
}
