package driver

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/go-sqlite/sqlitex"
)

var (
	DefaultPath string
)

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "./tmp"
	}

	DefaultPath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

const (
	FilePathSuffix             = ".sqlite"
	DefaultLockWaitTimeSeconds = 5
)

type Connection struct {
	Conn         *sqlite.Conn
	mu           *sync.Mutex
	DBID         string
	lock         LockType
	path         string
	readOnly     bool
	lockWaitTime time.Duration
	injectables  []*InjectableVar
	opts         []ConnOpt
}

// OpenConn opens a connection to the database with the given ID/name.
func OpenConn(dbid string, opts ...ConnOpt) (*Connection, error) {
	connection := &Connection{
		DBID:         dbid,
		mu:           &sync.Mutex{},
		lock:         LOCK_TYPE_UNLOCKED,
		path:         DefaultPath,
		readOnly:     false,
		lockWaitTime: time.Second * DefaultLockWaitTimeSeconds,
		opts:         opts,
	}
	for _, opt := range opts {
		opt(connection)
	}

	if err := connection.openConn(); err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	return connection, nil
}

func (c *Connection) openConn() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	flags := sqlite.OpenReadWrite | sqlite.OpenCreate
	if c.readOnly {
		flags = sqlite.OpenReadOnly
	}

	if c.Conn == nil {
		fp := c.getFilePath()
		err := createDirIfNeeded(fp)
		if err != nil {
			return err
		}

		conn, err := sqlite.OpenConn(fp, flags)
		if err != nil {
			return err
		}
		c.Conn = conn
		c.lock = LOCK_TYPED_SHARED
	}

	return nil
}

func (c *Connection) getFilePath() string {
	return c.path + c.DBID + FilePathSuffix
}

// Close closes the connection to the database
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.releaseLock()
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			return err
		}
		c.Conn = nil
	}
	return nil
}

// Execute executes a statement
func (c *Connection) Execute(stmt string, args ...interface{}) error {
	return c.execute(stmt, &sqlitex.ExecOptions{
		Args:          args,
		OverrideFlags: sqlitex.ForbidMissing,
	})
}

// ExecuteNamed executes a statement
func (c *Connection) ExecuteNamed(stmt string, args map[string]interface{}, resultFns ...ResultFn) error {
	if args == nil {
		args = make(map[string]interface{})
	}
	c.addInjectables(args)
	return c.execute(stmt, &sqlitex.ExecOptions{
		Named:         args,
		OverrideFlags: sqlitex.ForbidMissing,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			dStmt := newStatement(stmt)
			for _, resultFn := range resultFns {
				if err := resultFn(dStmt); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func (c *Connection) execute(stmt string, options *sqlitex.ExecOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.Writable() {
		return ErrNoWriteLock
	}

	stmt = trimPadding(stmt)
	if stmt == "" {
		return fmt.Errorf("statement is empty")
	}

	err := sqlitex.Execute(c.Conn, stmt, options)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

// a ResultFn is a function that is called for each row returned by a query
type ResultFn func(*Statement) error

// Query executes a query and calls the resultFn for each row returned
func (c *Connection) Query(statement string, resultFn ResultFn, args ...interface{}) error {
	return c.query(statement, resultFn, func(stmt *Statement) error {
		return stmt.BindMany(args)
	})
}

// QueryNamed executes a query and calls the resultFn for each row returned
func (c *Connection) QueryNamed(statement string, resultFn ResultFn, args map[string]interface{}) error {
	return c.query(statement, resultFn, func(stmt *Statement) error {
		c.addInjectables(args)
		return stmt.SetMany(args)
	})
}

// Prepare prepares a statement for execution and stores it in the connection
func (c *Connection) Prepare(statement string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.prepare(statement)
	return err
}

// query executes a query and calls the resultFn for each row returned
// statementSetterFn is a function that is called to set the arguments for the statement
func (c *Connection) query(statement string, resultFn ResultFn, statementSetterFn func(*Statement) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	statement = trimPadding(statement)
	stmt, err := c.prepare(statement)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Clear()

	if err := statementSetterFn(stmt); err != nil {
		return err
	}

	for {
		hasRow, err := stmt.step()
		if err != nil {
			return fmt.Errorf("failed to step statement: %w", err)
		}

		if !hasRow {
			break
		}

		c.mu.Unlock()
		err = resultFn(stmt)
		c.mu.Lock()

		if err != nil {
			return err
		}
	}

	return nil
}

// Prepare prepares a statement for execution
func (c *Connection) prepare(statement string, extraParams ...string) (*Statement, error) {
	if !c.Readable() {
		return nil, ErrConnectionClosed
	}

	sqliteStmt, err := c.Conn.Prepare(statement, c.listInjectables()...)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return newStatement(sqliteStmt), nil
}

func (c *Connection) AcquireLock() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := acquireLock(c.DBID, c.lockWaitTime)
	if err != nil {
		return err
	}

	c.lock = LOCK_TYPE_EXCLUSIVE
	return nil
}

func (c *Connection) ReleaseLock() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.releaseLock()
}

func (c *Connection) releaseLock() {
	if c.lock != LOCK_TYPE_EXCLUSIVE {
		return
	}

	releaseLock(c.DBID)
	c.lock = LOCK_TYPED_SHARED
}

func (c *Connection) LastInsertRowID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Conn.LastInsertRowID()
}

func trimPadding(s string) string {
	return strings.TrimSpace(s)
}

func (c *Connection) DisableForeignKeys() error {
	return c.Execute("PRAGMA foreign_keys = OFF;")
}

func (c *Connection) EnableForeignKeys() error {
	return c.Execute("PRAGMA foreign_keys = ON;")
}

func (c *Connection) CopyReadOnly() (*Connection, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn == nil {
		return nil, ErrConnectionClosed
	}

	newOpts := make([]ConnOpt, 0)
	newOpts = append(newOpts, c.opts...)
	newOpts = append(newOpts, ReadOnly())
	return OpenConn(c.DBID, newOpts...)
}
