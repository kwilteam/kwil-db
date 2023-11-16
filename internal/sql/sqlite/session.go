package sqlite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/go-sqlite/sqlitex"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// ApplyChangeset applies a changeset to the connection.
// It will return an error if the connection is read-only.
func (c *Connection) ApplyChangeset(reader io.Reader) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReadonly() {
		return ErrReadOnlyConn
	}

	return c.conn.ApplyChangeset(reader, nil, func(ct sqlite.ConflictType, ci *sqlite.ChangesetIterator) sqlite.ConflictAction {
		op, err := ci.Operation()
		if err != nil {
			return sqlite.ChangesetAbort
		}

		switch op.Type {
		case sqlite.OpInsert:
			return sqlite.ChangesetReplace
		case sqlite.OpDelete:
			return sqlite.ChangesetOmit
		case sqlite.OpUpdate:
			return sqlite.ChangesetReplace
		default:
			return sqlite.ChangesetAbort
		}
	})
}

// listTables lists all tables (including metadata tables) in the database.
func (c *Connection) listTables() ([]string, error) {
	tables := make([]string, 0)
	err := sqlitex.ExecuteTransient(c.conn, sqlListTables, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			tables = append(tables, stmt.ColumnText(0))
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf(`failed to execute "list tables" query: %w`, err)
	}

	return tables, nil
}

// getColumNames gets the column names for the given table.
func (c *Connection) getColumnNames(ctx context.Context, table string) ([]string, error) {
	res, err := c.execute(ctx, "PRAGMA table_info(users)", nil)
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	columns := []string{}
	resCols := res.Columns()
	nameIdx := 0
	for i, col := range resCols {
		if col == "name" {
			nameIdx = i
		}
	}
	for {
		rowReturned, err := res.Next()
		if err != nil {
			return nil, err
		}

		if !rowReturned {
			break
		}

		vals, err := res.Values()
		if err != nil {
			return nil, err
		}

		columns = append(columns, vals[nameIdx].(string))
	}

	return columns, nil
}

func filterPublicTables(tables []string) []string {
	filtered := make([]string, 0)
	for _, table := range tables {
		if table == "sqlite_master" || table[0] == '_' {
			continue
		}

		filtered = append(filtered, table)
	}

	return filtered
}

func (c *Connection) CreateSession() (sql.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReadonly() {
		return nil, ErrReadOnlyConn
	}

	// only track tables that are not "sqlite_master" or start with "_"
	tables, err := c.listTables()
	if err != nil {
		return nil, err
	}

	ses, err := c.conn.CreateSession("")
	if err != nil {
		return nil, err
	}

	for _, table := range filterPublicTables(tables) {
		err = ses.Attach(table)
		if err != nil {
			ses.Delete()
			return nil, err
		}
	}

	return &Session{
		session: ses,
		conn:    c,
	}, nil
}

type Session struct {
	mu      sync.Mutex
	session *sqlite.Session
	conn    *Connection
}

// Delete deletes the session and associated resources.
func (s *Session) Delete() (err error) {
	defer func() {
		// recover from panic if one occurred. Set err to nil otherwise.
		if r := recover(); r != nil {
			err = fmt.Errorf("error closing session: %v", r)
		}
	}()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.Delete()

	return nil
}

// ChangesetID returns the changeset ID for the session.
// It is a deterministic identifier based on the changed data.
func (s *Session) ChangesetID(ctx context.Context) ([]byte, error) {
	cs, err := s.Changeset(ctx)
	if err != nil {
		return nil, err
	}

	return cs.ID()
}

// Changeset returns the changeset for the session.
func (s *Session) Changeset(ctx context.Context) (*Changeset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := new(bytes.Buffer)
	err := s.session.WritePatchset(buf)
	if err != nil {
		return nil, err
	}

	iter, err := sqlite.NewChangesetIterator(buf)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	return s.createChangeset(ctx, iter)
}
