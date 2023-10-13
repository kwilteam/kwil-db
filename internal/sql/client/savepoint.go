package client

import (
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
)

type Savepoint struct {
	sp *sqlite.Savepoint
}

func (s *Savepoint) Rollback() error {
	return s.sp.Rollback()
}

func (s *Savepoint) Commit() error {
	return s.sp.Commit()
}

func (s *Savepoint) CommitAndCheckpoint() error {
	return s.sp.CommitAndCheckpoint()
}
