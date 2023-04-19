package driver

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/kwilteam/go-sqlite"
	"go.uber.org/zap"
)

type Savepoint struct {
	*Connection
	name       string
	ses        *sqlite.Session
	sesDeleted bool
	committed  bool
}

// Creates a savepoint with the given name. If no name is provided, a random
// name will be generated.
// Only the first argument is used, so you can pass in a string literal
func (c *Connection) Savepoint(nameArr ...string) (*Savepoint, error) {
	var name string
	// generate a random name if none is provided
	if len(nameArr) == 0 {
		name = randomSavepointName(10)
	} else {
		name = nameArr[0]
	}

	c.log.Debug("Creating savepoint", zap.String("name", name))

	if err := c.Execute("SAVEPOINT " + name); err != nil {
		return nil, err
	}

	ses, err := c.Conn.CreateSession("")
	if err != nil {
		return nil, err
	}

	err = ses.Attach("")
	if err != nil {
		return nil, err
	}

	return &Savepoint{
		Connection: c,
		name:       name,
		ses:        ses,
		sesDeleted: false,
		committed:  false,
	}, nil
}

// end cleans up anything that needs to be closed
// after the savepoint is committed or rolled back
func (s *Savepoint) end() {
	if !s.sesDeleted {
		s.ses.Delete()
		s.sesDeleted = true
	}
}

// Commit commits or "releases" the savepoint.
// I use the term commit since it is more clear for most devs,
// but the technical term for SQLite is "release".
func (s *Savepoint) Commit() error {
	if s.committed {
		return fmt.Errorf("commit failed: savepoint already committed or rolled back")
	}
	s.committed = true
	defer s.end()
	return s.Execute("RELEASE " + s.name)
}

func (s *Savepoint) Rollback() error {
	if s.committed {
		return fmt.Errorf("rollback failed: savepoint already committed or rolled back")
	}
	s.committed = true

	defer s.end()
	s.log.Debug("Rolling back savepoint", zap.String("name", s.name))
	return s.Execute("ROLLBACK TO " + s.name)
}

// GetChangeset returns a bytes.Buffer containing the changeset
func (s *Savepoint) GetChangeset() (*bytes.Buffer, error) {
	err := s.ses.Diff("main", "main")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := io.Writer(&buf)

	err = s.ses.WriteChangeset(writer)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// ApplyChangeset applies the changeset to the database.
// If it fails, it will return an error.
func (s *Savepoint) ApplyChangeset(changeset *bytes.Buffer) error {
	reader := io.Reader(changeset)

	return s.Conn.ApplyChangeset(reader, nil, func(ct sqlite.ConflictType, ci *sqlite.ChangesetIterator) sqlite.ConflictAction {
		return sqlite.ChangesetAbort
	})
}

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var alphanumericRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomSavepointName(length int) string {
	if length < 2 {
		panic("Length must be at least 2 to generate a valid savepoint name.")
	}

	result := make([]rune, length)
	// First character must be a letter
	result[0] = letterRunes[rand.Intn(len(letterRunes))]

	// Rest of the characters can be alphanumeric
	for i := 1; i < length; i++ {
		result[i] = alphanumericRunes[rand.Intn(len(alphanumericRunes))]
	}

	return string(result)
}
