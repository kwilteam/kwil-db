package sqlite

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type SqliteStore struct {
	conn         *sqlite.Connection
	metadataStmt *sqlite.Statement
	log          log.Logger
	path         string
	name         string
	globalVars   map[string]any
}

func NewSqliteStore(name string, opts ...SqliteOpts) (*SqliteStore, error) {
	sqliteDB := &SqliteStore{
		log:  log.NewNoOp(),
		name: name,
		path: defaultPath,
	}

	for _, opt := range opts {
		opt(sqliteDB)
	}

	var err error
	sqliteDB.conn, err = sqlite.OpenConn(sqliteDB.name,
		sqlite.WithPath(sqliteDB.path),
		sqlite.WithLogger(sqliteDB.log),
		sqlite.WithGlobalVariables(sqliteDB.globalVars),
	)
	if err != nil {
		return nil, err
	}

	err = sqliteDB.initTables()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return sqliteDB, nil
}

func (s *SqliteStore) Prepare(query string) (sqldb.Statement, error) {
	sqliteStmtString, err := parseSql(query)
	if err != nil {
		return nil, err
	}

	return s.PrepareRaw(sqliteStmtString)
}

func (s *SqliteStore) Query(ctx context.Context, query string, args map[string]any) (dto.Result, error) {
	res := &sqlite.ResultSet{}

	err := s.conn.Query(ctx, query,
		sqlite.WithNamedArgs(args),
		sqlite.WithResultSet(res),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *SqliteStore) Close() error {
	ch := make(chan struct{})
	err := s.conn.Close(ch)
	if err != nil {
		return err
	}

	<-ch

	return nil
}

func (s *SqliteStore) Delete() error {
	return s.conn.Delete()
}

func (s *SqliteStore) Savepoint() (sqldb.Savepoint, error) {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return nil, err
	}

	return newSavepoint(sp), nil
}

// PrepareRaw is like Prepare, but does not parse the query.
func (s *SqliteStore) PrepareRaw(stmt string) (sqldb.Statement, error) {
	sqliteStmt, err := s.conn.Prepare(stmt)
	if err != nil {
		return nil, err
	}

	return NewStatement(sqliteStmt), nil
}

// Execute executes a statement
func (s *SqliteStore) Execute(stmt string, inputs map[string]any) error {
	return s.conn.Execute(stmt, inputs)
}
