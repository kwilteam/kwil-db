package sqlite

import (
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type Savepoint struct {
	sp *sqlite.Savepoint
}

func newSavepoint(sp *sqlite.Savepoint) *Savepoint {
	return &Savepoint{sp: sp}
}

func (s *Savepoint) Rollback() error {
	return s.sp.Rollback()
}

func (s *Savepoint) Commit() error {
	return s.sp.Commit()
}
