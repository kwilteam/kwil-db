package sqlite

import (
	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type Statement struct {
	stmt *sqlite.Statement
}

func NewStatement(stmt *sqlite.Statement) *Statement {
	return &Statement{stmt: stmt}
}

func (s *Statement) Execute(args map[string]any) (dto.Result, error) {
	res := &sqlite.ResultSet{}

	err := s.stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: args,
		ResultSet: res,
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Statement) Close() error {
	return s.stmt.Finalize()
}
