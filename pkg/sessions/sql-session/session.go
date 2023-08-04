package sqlsession

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"go.uber.org/zap"
)

type SqlCommitable struct {
	db SqlDB

	session   sql.Session
	savepoint sql.Savepoint

	log log.Logger
}

// NewSqlCommitable creates a new SqlCommitable.
func NewSqlCommitable(db SqlDB, opts ...SqlCommittableOpt) *SqlCommitable {
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
func (s *SqlCommitable) EndCommit(ctx context.Context, appender func([]byte) error) (commitId []byte, err error) {
	if s.session == nil {
		return nil, fmt.Errorf("session not started")
	}
	if s.savepoint == nil {
		return nil, fmt.Errorf("savepoint not active")
	}

	defer s.savepoint.Rollback()

	changes, err := s.session.GenerateChangeset()
	if err != nil {
		return nil, err
	}

	id, err := changes.ID()
	if err != nil {
		return nil, err
	}

	data, err := changes.Export()
	if err != nil {
		return nil, err
	}

	err = appender(data)
	if err != nil {
		return nil, err
	}

	errs := []error{}
	err = changes.Close()
	if err != nil {
		errs = append(errs, err)
	}

	err = s.session.Delete()
	if err != nil {
		errs = append(errs, err)
	}

	s.session = nil

	return id, errors.Join(errs...)
}

// BeginApply starts a new savepoint for applying changes.
func (s *SqlCommitable) BeginApply(ctx context.Context) error {
	if s.savepoint != nil {
		return fmt.Errorf("savepoint already active")
	}

	savepoint, err := s.db.Savepoint()
	if err != nil {
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
func (s *SqlCommitable) Cancel(ctx context.Context) {
	errs := []error{}
	if s.session != nil {
		errs = append(errs, s.session.Delete())
		s.session = nil
	}
	if s.savepoint != nil {
		errs = append(errs, s.savepoint.Rollback())
		s.savepoint = nil
	}

	if len(errs) > 0 {
		s.log.Error("errors while cancelling session", zap.Error(errors.Join(errs...)))
	}
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
}
