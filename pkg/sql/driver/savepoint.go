package driver

import (
	"bytes"
	"io"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const mainDB = "main"

type savepoint struct {
	conn       *sqlite.Conn
	notifyChan chan error // notifies when a savepoint is committed. an error can be sent to this channel to rollback the savepoint
	session    *sqlite.Session
	Started    bool
}

func newSavepoint(conn *sqlite.Conn) *savepoint {
	return &savepoint{
		conn:       conn,
		notifyChan: make(chan error),
		session:    nil,
		Started:    false,
	}
}

// Start will start a savepoint and return an error if a savepoint is already active.
func (s *savepoint) Start() (err error) {
	if s.Started {
		return ErrActiveSavepoint
	}
	s.session, err = s.conn.CreateSession("")
	if err != nil {
		return err
	}
	err = s.session.Attach("")
	if err != nil {
		return err
	}

	go func() {
		err2, ok := <-s.notifyChan
		if !ok {
			panic("savepoint channel closed unexpectedly, or was never opened")
		}
		sqlitex.Save(s.conn)(&err2)
	}()

	s.Started = true
	return nil
}

// Commit will commit the savepoint and close the savepoint notify channel
func (s *savepoint) Commit() error {
	return s.end(nil)
}

// Rollback will rollback the savepoint and close the savepoint notify channel
func (s *savepoint) Rollback() error {
	return s.end(ErrSavepointRollback)
}

// end will close the savepoint notify channel and set the savepoint as not started.
func (s *savepoint) end(err error) error {
	if !s.Started {
		return ErrNoActiveSavepoint
	}

	s.notifyChan <- err

	s.Started = false
	s.safeDeleteSession()
	return nil
}

func (s *savepoint) safeDeleteSession() {
	if s.session != nil {
		s.session.Delete()
		s.session = nil
	}
}

// GenerateChangeset will generate a changeset from the savepoint.
func (s *savepoint) GenerateChangeset() (*bytes.Buffer, error) {
	err := s.session.Diff(mainDB, mainDB)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := io.Writer(&buf)

	err = s.session.WriteChangeset(writer)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// Savepoint begins a new savepoint.
// If there is already a savepoint active, this will return an error.
func (c *Connection) BeginSavepoint() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.savepoint == nil {
		c.savepoint = newSavepoint(c.conn)
	}

	// this will catch if there is already a savepoint active
	if err := c.savepoint.Start(); err != nil {
		return err
	}

	return nil
}

// CommitSavepoint will commit the current savepoint.
// If there is no savepoint active, this will return an error.
func (c *Connection) CommitSavepoint() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.savepoint == nil {
		return ErrNoActiveSavepoint
	}

	return c.savepoint.Commit()
}

// RollbackSavepoint will rollback the current savepoint.
// If there is no savepoint active, this will return an error.
func (c *Connection) RollbackSavepoint() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.savepoint == nil {
		return ErrNoActiveSavepoint
	}

	return c.savepoint.Rollback()
}

// ActiveSavepoint returns true if there is an active savepoint.
func (c *Connection) ActiveSavepoint() bool {
	return c.savepoint.Started
}

// GenerateChangeset will generate a changeset from the current savepoint.
// If there is no savepoint active, this will return an error.
func (c *Connection) GenerateChangeset() (*bytes.Buffer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.savepoint == nil {
		return nil, ErrNoActiveSavepoint
	}

	return c.savepoint.GenerateChangeset()
}
