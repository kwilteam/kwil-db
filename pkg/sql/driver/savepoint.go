package driver

import (
	"bytes"
	"io"
	"math/rand"
	"time"

	"zombiezen.com/go/sqlite"
)

type Savepoint struct {
	*Connection
	name string
	ses  *sqlite.Session
}

// Creates a savepoint with the given name. If no name is provided, a random
// name will be generated.
// Only the first argument is used, so you can pass in a string literal
func (c *Connection) Savepoint(nameArr ...string) (*Savepoint, error) {
	var name string
	// generate a random name if none is provided
	if len(nameArr) == 0 {
		name = randomString(10)
	} else {
		name = nameArr[0]
	}

	if err := c.Execute("SAVEPOINT " + name); err != nil {
		return nil, err
	}

	ses, err := c.conn.CreateSession("")
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
	}, nil
}

// end cleans up anything that needs to be closed
// after the savepoint is committed or rolled back
func (s *Savepoint) end() {
	s.ses.Disable()
}

// Commit commits or "releases" the savepoint.
// I use the term commit since it is more clear for most devs,
// but the technical term for SQLite is "release".
func (s *Savepoint) Commit() error {
	defer s.end()
	return s.Execute("RELEASE " + s.name)
}

func (s *Savepoint) Rollback() error {
	defer s.end()
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

	return s.conn.ApplyChangeset(reader, nil, func(ct sqlite.ConflictType, ci *sqlite.ChangesetIterator) sqlite.ConflictAction {
		return sqlite.ChangesetAbort
	})
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
