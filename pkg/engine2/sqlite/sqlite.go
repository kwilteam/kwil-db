package sqlite

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	sqlitegenerator "github.com/kwilteam/kwil-db/pkg/engine2/sqlite/sql-ddl-generator"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type SqliteDB struct {
	conn         *sqlite.Connection
	metadataStmt *sqlite.Statement
}

func NewSqliteDB(conn *sqlite.Connection) (*SqliteDB, error) {
	sqliteDB := &SqliteDB{
		conn: conn,
	}

	err := sqliteDB.initTables()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return sqliteDB, nil
}

func (s *SqliteDB) Prepare(query string) (*Statement, error) {
	sqliteStmtString, err := parseSql(query)
	if err != nil {
		return nil, err
	}

	stmt, err := s.conn.Prepare(sqliteStmtString)
	if err != nil {
		return nil, err
	}

	return NewStatement(stmt), nil
}

func (s *SqliteDB) Query(ctx context.Context, query string, args map[string]any) (dto.Result, error) {
	res := &sqlite.ResultSet{}

	err := s.conn.Query(ctx, query, &sqlite.ExecOpts{
		NamedArgs: args,
		ResultSet: res,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *SqliteDB) Close() error {
	return s.conn.Close(nil)
}

func (s *SqliteDB) Savepoint() (*Savepoint, error) {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return nil, err
	}

	return NewSavepoint(sp), nil
}

// deployTable deploys a table to the database.
// If the table already exists, it is not deployed, and false is returned.
// If the table does not exist, it is deployed, and true is returned.
func (s *SqliteDB) deployTable(ctx context.Context, table *dto.Table) (bool, error) {
	exists, err := s.conn.TableExists(ctx, table.Name)
	if err != nil {
		return false, fmt.Errorf("failed to check if table %s exists: %w", table.Name, err)
	}

	if exists {
		return false, nil
	}

	stmts, err := sqlitegenerator.GenerateDDL(table)
	if err != nil {
		return false, fmt.Errorf("failed to generate DDL for table %s: %w", table.Name, err)
	}

	for _, stmt := range stmts {
		err := s.conn.Execute(stmt)
		if err != nil {
			return false, fmt.Errorf("failed to execute DDL for table %s: %w", table.Name, err)
		}
	}

	return true, nil
}

func (s *SqliteDB) CreateTable(ctx context.Context, table *dto.Table) error {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	alreadyExists, err := s.deployTable(ctx, table)
	if err != nil {
		return err
	}

	if alreadyExists {
		return fmt.Errorf("table %s already exists", table.Name)
	}

	err = s.storeTable(ctx, table)
	if err != nil {
		return err
	}

	return sp.Commit()
}

func (s *SqliteDB) CreateAction(ctx context.Context, action *dto.Action) error {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	err = s.storeAction(ctx, action)
	if err != nil {
		return err
	}

	return sp.Commit()
}
