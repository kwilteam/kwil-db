package sqlite

import (
	"context"
	"fmt"
	"kwil/pkg/log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/go-sqlite/sqlitex"
	"go.uber.org/zap"
)

type Connection struct {
	conn            *sqlite.Conn
	mu              *sync.Mutex // mutex to protect the write connection
	globalVariables []*GlobalVariable
	log             log.Logger
	path            string
	readPool        *sqlitex.Pool
	poolSize        int
	flags           sqlite.OpenFlags
	isMemory        bool
	name            string
}

func OpenConn(name string, opts ...ConnectionOption) (*Connection, error) {
	connection := &Connection{
		log:      log.NewNoOp(),
		mu:       &sync.Mutex{},
		path:     DefaultPath,
		name:     name,
		poolSize: 10,
		conn:     nil,
		readPool: nil,
		isMemory: false,
		flags:    sqlite.OpenWAL,
	}
	for _, opt := range opts {
		opt(connection)
	}
	if !connection.isMemory {
		err := connection.mkPathDir()
		if err != nil {
			return nil, fmt.Errorf("failed to create path dir: %w", err)
		}
	}

	err := connection.openConn()
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	return connection, nil
}

func (c *Connection) openFlags(readOnly bool) sqlite.OpenFlags {
	if readOnly {
		return c.flags | sqlite.OpenReadOnly
	}
	return c.flags | sqlite.OpenReadWrite | sqlite.OpenCreate
}

func (c *Connection) getFilePath() string {
	if c.isMemory {
		return c.path
	}
	return fmt.Sprintf("%s%s.sqlite", c.path, c.name)
}

func (c *Connection) openConn() error {
	var err error
	c.conn, err = sqlite.OpenConn(c.getFilePath(), c.openFlags(false))
	if err != nil {
		return fmt.Errorf("failed to open readwrite connection: %w", err)
	}

	c.readPool, err = sqlitex.Open(c.getFilePath(), c.openFlags(true), c.poolSize)
	if err != nil {
		return fmt.Errorf("failed to create read connection pool: %w", err)
	}

	return nil
}

func (c *Connection) mkPathDir() error {
	dir := filepath.Dir(c.path)
	return os.MkdirAll(dir, os.ModePerm)
}

// execute executes a statement on the write connection.
// dml statements should use prepared statements instead unless they are one-offs.
// this method is intentionally barebones to prevent misuse.
func (c *Connection) Execute(stmt string) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.execute(stmt)
}

// execute executes a one-off statement.  It does not use a mutex, unlike Execute.
func (c *Connection) execute(stmt string) error {
	return sqlitex.ExecuteTransient(c.conn, stmt, nil)
}

func (c *Connection) Prepare(stmt string) (*Statement, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	innerStmt, err := c.conn.Prepare(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return newStatement(c, innerStmt), nil
}

// Close closes the connection.
// It takes an optional wait channel, which will be waited on before the connection is closed.
func (c *Connection) Close(ch chan<- struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.close(ch)
}

// close closes the connection.
// It takes an optional wait channel, which will be waited on before the connection is closed.
func (c *Connection) close(ch chan<- struct{}) error {
	if c.conn == nil { // if the connection is nil, it's already closed / been deleted
		return nil
	}

	go func(ch chan<- struct{}) {
		err := c.readPool.Close()
		if err != nil {
			c.log.Error("failed to close read connection pool", zap.Error(err))
		}

		err = c.conn.Close()
		if err != nil {
			c.log.Error("failed to close readwrite connection", zap.Error(err))
		}

		if ch != nil {
			ch <- struct{}{}
		}
	}(ch)

	return nil
}

// prepareRead prepares a read-only query.  it returns a statement, a deferFunc which returns the connection to the pool and finalizes the statement, and an error.
func (c *Connection) prepareRead(ctx context.Context, statement string) (stmt *Statement, deferFunc func() error, err error) {
	readConn := c.readPool.Get(ctx)
	if readConn == nil {
		return nil, nil, fmt.Errorf("failed to get read connection from connection pool")
	}

	deferFunc = func() error {
		c.readPool.Put(readConn)
		return stmt.stmt.Finalize()
	}

	innerStmt, trailingBytes, err := readConn.PrepareTransient(trimPadding(statement))
	if err != nil {
		c.readPool.Put(readConn)
		return nil, deferFunc, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if trailingBytes > 0 {
		deferFunc()
		return nil, deferFunc, fmt.Errorf("trailing bytes after statement: %q", trailingBytes)
	}

	return newStatement(c, innerStmt), deferFunc, nil
}

// Query executes a read-only query against the database.
// It takes a QueryOpts struct, which can contain arguments, a function to manually bind parameters, a function
// to manually handle each result in between Step() calls, and a struct to store the results in.
// All of these are optional, and if not provided, the function will return an error
func (c *Connection) Query(ctx context.Context, statement string, opts *ExecOpts) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if opts == nil {
		return fmt.Errorf("query options cannot be nil")
	}
	if opts.ResultFunc == nil {
		opts.ResultFunc = func(*Statement) error { return nil }
	}

	err := c.query(ctx, statement, opts)
	if err != nil {
		c.log.Error("failed to execute query", zap.Error(err))
	}

	return nil
}

// query executes a query and calls the resultFn for each row returned
// statementSetterFn is a function that is called to set the arguments for the statement
func (c *Connection) query(ctx context.Context, statement string, opts *ExecOpts) error {
	stmt, deferFunc, err := c.prepareRead(ctx, statement)
	if err != nil {
		return fmt.Errorf("error preparing read: %w", err)
	}
	defer deferFunc()

	return stmt.Execute(opts)
}

func (c *Connection) CheckpointWal() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.execute(sqlCheckpoint)
}

func (c *Connection) EnableForeignKey() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.execute(sqlEnableFK)
}

func (c *Connection) DisableForeignKey() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.execute(sqlDisableFK)
}

/*
func (c *Connection) TableExistsW(ctx context.Context, tableName string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

}*/

func (c *Connection) ListTables(ctx context.Context) ([]string, error) {
	tables := make([]string, 0)
	err := c.Query(ctx, sqlListTables, &ExecOpts{
		ResultFunc: func(stmt *Statement) error {
			tables = append(tables, stmt.GetText("name"))
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	return tables, nil
}

func (c *Connection) TableExists(ctx context.Context, tableName string) (bool, error) {
	exists := false
	err := c.Query(ctx, sqlIfTableExists, &ExecOpts{
		NamedArgs: map[string]interface{}{
			"$name": tableName,
		},
		ResultFunc: func(stmt *Statement) error {
			exists = true
			return nil
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if table exists: %w", err)
	}

	return exists, nil
}

// Delete deletes the entire database, and sets the underlying connection to nil.
// It waits for all read connections to close before deleting the database.
func (c *Connection) Delete() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	waitChan := make(chan struct{})
	err := c.close(waitChan)
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	<-waitChan

	if c.isMemory {
		c.conn = nil
		c.readPool = nil
		return nil
	}

	err = c.deleteFiles()
	if err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	c.conn = nil
	c.readPool = nil

	return nil
}

func (c *Connection) deleteFiles() error {
	if c.isMemory {
		return nil
	}

	files := []string{
		c.getFilePath(),
		fmt.Sprintf("%s-wal", c.getFilePath()),
		fmt.Sprintf("%s-shm", c.getFilePath()),
	}

	for _, file := range files {
		err := deleteIfExists(file)
		if err != nil {
			return fmt.Errorf("failed to delete file %q: %w", file, err)
		}
	}

	return nil
}

func deleteIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

func trimPadding(s string) string {
	ss := strings.TrimSpace(s)
	return ss
}

type ResultSet struct {
	Rows    [][]any  `json:"rows"`
	Columns []string `json:"columns"`
	index   int
}

func (r *ResultSet) Next() bool {
	r.index++
	return r.index < len(r.Rows)
}
