package sqlite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/go-sqlite/sqlitex"
	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite/functions"
)

var (
	globalMu sync.Mutex                  // globalMu protects the global map of open databases.
	openDBs  = make(map[string]struct{}) // Map to hold DB names that have an open writer
)

// Open opens a connection to a sqlite database.
// It will open with foreign key constraints enabled.
func Open(ctx context.Context, name string, flags sql.ConnectionFlag) (*Connection, error) {
	// Fail defines extra actions to perform on an error path.
	fail := func(err error) error { return err } // passthrough to start

	globalMu.Lock()
	defer globalMu.Unlock()

	sqliteFlags := sqlite.OpenFlags(0)
	if flags&sql.OpenReadOnly != 0 {
		sqliteFlags |= sqlite.OpenReadOnly
	} else {
		sqliteFlags |= sqlite.OpenReadWrite

		// If the database is already open, return an error
		if _, ok := openDBs[name]; ok {
			return nil, ErrWriterOpen
		}

		// only if the database is RW and not in-memory do we want to track it
		if flags&sql.OpenMemory == 0 {
			openDBs[name] = struct{}{}
			fail = func(err error) error {
				delete(openDBs, name)
				return err
			}
		}
	}

	if flags&sql.OpenMemory != 0 {
		sqliteFlags |= sqlite.OpenMemory
	} else {
		sqliteFlags |= sqlite.OpenWAL
	}

	if flags&sql.OpenCreate != 0 {
		sqliteFlags |= sqlite.OpenCreate
	}

	// Extract the directory name from the file path
	dirName := filepath.Dir(name)

	// Create the directory along with any necessary parent directories
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return nil, fail(err)
	}

	done := make(chan struct{})
	var conn *sqlite.Conn
	var err error

	go func() {
		conn, err = sqlite.OpenConn(name, sqliteFlags)
		close(done)
	}()

	select {
	case <-ctx.Done():
		// Cancel ongoing operations here and then return
		return nil, fail(ctx.Err())
	case <-done:
		// Continue as normal
	}
	if err != nil {
		return nil, fail(err)
	}

	// If we fail now, we must also close the connection.
	{
		baseFail := fail // for capture in new fail
		fail = func(err error) error {
			return errors.Join(conn.Close(), baseFail(err))
		}
	}

	err = functions.Register(conn)
	if err != nil {
		return nil, fail(err)
	}

	err = conn.SetDefensive(true)
	if err != nil {
		return nil, fail(err)
	}

	c := &Connection{
		closed: make(chan struct{}),
		conn:   conn,
		flags:  flags,
		file:   name,
	}
	// NOTE:  We must not use the Connection's close method since it tries to
	// acquire globalMu; we can do the cleanup directly.

	if !c.isReadonly() {
		err = c.EnableForeignKey()
		if err != nil {
			return nil, fail(err)
		}
	}

	res, err := c.execute(ctx, sqlPragmaSync, nil)
	if err != nil {
		return nil, fail(err)
	}
	err = res.Finish()
	if err != nil {
		return nil, fail(err)
	}

	err = initializeKv(ctx, c)
	if err != nil {
		return nil, fail(err)
	}

	return c, nil
}

// connection is a single connection to a sqlite database.
type Connection struct {
	mu     sync.Mutex // mu protects all exported methods of Connection
	closed chan struct{}
	conn   *sqlite.Conn
	flags  sql.ConnectionFlag
	file   string

	// inUse is true if the connection is in use.
	inUse atomic.Bool
	// activeSavepoint is true if there is an active savepoint.
	// if false, then a new savepoint will checkpoint the wal when committed.
	activeSavepoint atomic.Bool
}

// Execute executes a statement against the connection.
// This can be either a write or read query.
// If a write is executed against a read-only connection, an error will be returned.
func (c *Connection) Execute(ctx context.Context, stmt string, args map[string]any) (sql.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.execute(ctx, stmt, args)
}

// execute executes a statement.
// it does not acquire the connection mutex.
func (c *Connection) execute(ctx context.Context, stmt string, args map[string]any) (res sql.Result, err error) {
	if c.inUse.Swap(true) {
		return nil, ErrInUse
	}
	defer func() {
		if err != nil {
			c.inUse.Store(false)
		}
	}()

	c.conn.SetBlockOnBusy()
	c.conn.SetInterrupt(ctx.Done())

	runningStmt := c.conn.CheckReset()
	if runningStmt != "" {
		return nil, fmt.Errorf("connection is busy with statement %q", runningStmt)
	}

	prepared, err := c.conn.Prepare(cleanStmt(stmt))
	if err != nil {
		return nil, err
	}

	err = setMany(prepared, args)
	if err != nil {
		return nil, err
	}

	columns := determineColumnNames(prepared)

	r := &Result{
		stmt:        prepared,
		columnNames: columns,
		closeFn: sync.OnceFunc(func() {
			c.inUse.Store(false)
			c.conn.SetInterrupt(nil)
		}),
		conn: c,
	}

	return r, nil
}

func cleanStmt(stmt string) string {
	trimmed := strings.TrimSpace(stmt)

	// check if ends with semicolon
	// if not, add one
	if !strings.HasSuffix(trimmed, ";") {
		return trimmed + ";"
	}
	return trimmed
}

// Close closes the connection.
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.close()
}

// close closes the connection.
// it does not acquire the connection mutex.
func (c *Connection) close() error {
	if c.isClosed() {
		return nil
	}

	err := c.conn.Close()
	if err != nil {
		return err
	}

	globalMu.Lock()
	if !c.isReadonly() && !c.isMemory() {
		delete(openDBs, c.file)
	}
	globalMu.Unlock()

	close(c.closed)

	return nil
}

func (c *Connection) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

// Delete deletes the database file.
// If the connection is read-only, an error will be returned.
func (c *Connection) DeleteDatabase() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReadonly() {
		return ErrReadOnlyConn
	}

	err := c.close()
	if err != nil {
		return err
	}

	if c.isMemory() {
		return nil
	}

	files := []string{
		c.file,
		fmt.Sprintf("%s-wal", c.file),
		fmt.Sprintf("%s-shm", c.file),
	}

	for _, file := range files {
		_, err = os.Stat(c.file)

		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to stat file %q: %w", file, err)
		}

		err = os.Remove(c.file)
		if err != nil {
			return fmt.Errorf("failed to delete file %q: %w", file, err)
		}
	}

	return nil
}

// isReadonly returns whether the connection is read-only.
func (c *Connection) isReadonly() bool {
	return c.flags&sql.OpenReadOnly != 0
}

// isMemory returns whether the connection is in-memory.
func (c *Connection) isMemory() bool {
	return c.flags&sql.OpenMemory != 0
}

// checkpointWal checkpoints the wal.
// If the connection is read-only, an error will be returned.
func (c *Connection) checkpointWal() error {

	if c.isReadonly() {
		return ErrReadOnlyConn
	}

	return execute(c.conn, sqlCheckpoint)
}

// EnableForeignKey enables foreign key constraints.
// If the connection is read-only, an error will be returned.
func (c *Connection) EnableForeignKey() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReadonly() {
		return ErrReadOnlyConn
	}

	return execute(c.conn, sqlEnableFK)
}

// DisableForeignKey disables foreign key constraints.
// If the connection is read-only, an error will be returned.
func (c *Connection) DisableForeignKey() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReadonly() {
		return ErrReadOnlyConn
	}

	return execute(c.conn, sqlDisableFK)
}

// TableExists returns whether the table exists.
func (c *Connection) TableExists(ctx context.Context, table string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	exists := false
	err := sqlitex.ExecuteTransient(c.conn, sqlIfTableExists, &sqlitex.ExecOptions{
		Named: map[string]interface{}{
			"$name": table,
		},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			exists = true
			return nil
		},
	})
	if err != nil {
		return false, fmt.Errorf(`failed to execute "if table exists" query: %w`, err)
	}

	return exists, nil
}

// execute executes an ad-hoc statement against the connection.
// it does not cache the statement.
func execute(c *sqlite.Conn, stmt string) error {
	return sqlitex.ExecuteTransient(c, trimPadding(stmt), nil)
}

// trimPadding removes unnecessary padding from a statement.
func trimPadding(s string) string {
	ss := strings.TrimSpace(s)
	return ss
}
