/*
Package Client provides a useful interface for SQLite databases.

It is essentially a convenience wrapper around pkg/engine/sql/sqlite.

In the future, this can likely just be a part of the sqlite package, however
that package is currently very stable and I don't want to break it.  This
package is also pretty experimental, so it's best to keep it separate for now.

The primary purpose of this package is to create a minimal interface for
interacting with SQLite databases.
*/

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type SqliteClient struct {
	// conn is the underlying connection to the SQLite database.
	conn *sqlite.Connection

	// log is self-explanatory.
	log log.Logger

	// path is the path to the SQLite database.
	path string

	// name is the name of the SQLite database file.  if it doesn't exist, it will be created.
	name string
}

func NewSqliteStore(name string, opts ...SqliteOpts) (*SqliteClient, error) {
	sqliteDB := &SqliteClient{
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
	)
	if err != nil {
		return nil, err
	}

	err = sqliteDB.conn.EnableForeignKey()
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return sqliteDB, nil
}

// Execute executes a statement.
func (s *SqliteClient) Execute(stmt string, args ...map[string]any) error {
	if len(args) > 1 {
		return fmt.Errorf("only one set of args is supported") // the sqlite engine can support multiple sets of args, but we don't need that functionality
	}

	return s.conn.Execute(stmt, args...)
}

// Query executes a query and returns the result.
func (s *SqliteClient) Query(ctx context.Context, query string, args ...map[string]any) (io.Reader, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("only one set of args is supported") // the sqlite engine can support multiple sets of args, but we don't need that functionality
	}
	if len(args) == 0 {
		args = append(args, nil)
	}

	res := &sqlite.ResultSet{}

	err := s.conn.Query(ctx, query,
		sqlite.WithNamedArgs(args[0]),
		sqlite.WithResultSet(res),
	)
	if err != nil {
		return nil, err
	}

	return resultsToReader(res)
}

// Prepare prepares a statement for execution, and returns a Statement.
func (s *SqliteClient) Prepare(stmt string) (*Statement, error) {
	sqliteStmt, err := s.conn.Prepare(stmt)
	if err != nil {
		return nil, err
	}

	return &Statement{
		stmt: sqliteStmt,
	}, nil
}

// TableExists checks if a table exists.
func (s *SqliteClient) TableExists(ctx context.Context, table string) (bool, error) {
	return s.conn.TableExists(ctx, table)
}

// Close closes the connection to the database.
func (s *SqliteClient) Close() error {
	return s.conn.Close()
}

func (s *SqliteClient) Delete() error {
	return s.conn.Delete()
}

func (s *SqliteClient) Savepoint() (*Savepoint, error) {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return nil, err
	}

	return &Savepoint{
		sp: sp,
	}, nil
}

func resultsToReader(res *sqlite.ResultSet) (io.Reader, error) {
	recordBytes, err := json.Marshal(res.Records())
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(recordBytes), nil
}

// ResultsfromReader reads the results from a reader and returns them as an array of maps.
// WARNING: this will not convert byte arrays to strings, so you will need to do that yourself.
func ResultsfromReader(reader io.Reader) ([]map[string]any, error) {
	// Declare an empty array of map
	var rows []map[string]interface{}

	// Decode the JSON stream
	dec := json.NewDecoder(reader)
	for {
		var row []map[string]interface{}
		if err := dec.Decode(&row); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if row == nil {
			continue
		}

		rows = append(rows, row...)
	}

	return rows, nil
}
