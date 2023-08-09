package client

import (
	// "github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type SqliteSession struct {
	sess *sqlite.Session
}

func (s *SqliteSession) GenerateChangeset() ([]byte, error) {
	return s.sess.GenerateChangesetBytes()
}

func (s *SqliteSession) Delete() error {
	return s.sess.Delete()
}
