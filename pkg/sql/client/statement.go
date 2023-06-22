package client

import (
	"io"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type Statement struct {
	stmt *sqlite.Statement
}

func (s *Statement) Execute(args map[string]any) (io.Reader, error) {
	res := &sqlite.ResultSet{}

	err := s.stmt.Execute(
		sqlite.WithNamedArgs(args),
		sqlite.WithResultSet(res),
	)

	if err != nil {
		return nil, err
	}

	return resultsToReader(res)
}

func (s *Statement) Close() error {
	return s.stmt.Finalize()
}
