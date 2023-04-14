package driver

import (
	"fmt"
	"log"
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
	closed       bool
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
		closed:       false,
	}
	for _, opt := range opts {
		opt(connection)
	}

	connection.mu.Lock()
	defer connection.mu.Unlock()

	if err := connection.openConn(); err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	return connection, nil
}

func (c *Connection) openConn() error {
	flags := sqlite.OpenReadWrite | sqlite.OpenCreate | sqlite.OpenWAL
	if c.readOnly {
		flags = sqlite.OpenReadOnly | sqlite.OpenWAL
	}

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

	return nil
}

func (c *Connection) getFilePath() string {
	return c.path + c.DBID + FilePathSuffix
}

// Close closes the connection to the database
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

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
	defer c.tryUnlock()

	if resultFn == nil {
		resultFn = func(*Statement) error { return nil }
	}

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

func (c *Connection) tryUnlock() {
	c.mu.TryLock()
	c.mu.Unlock()
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
	newConn, err := OpenConn(c.DBID, newOpts...)
	if err != nil {
		return nil, err
	}

	go newConn.pollReOpen()

	return newConn, nil
}

const sqlIfTableExists = `SELECT name FROM sqlite_master WHERE type='table' AND name=$name;`

func (c *Connection) TableExists(name string) (bool, error) {
	exists := false
	err := c.QueryNamed(sqlIfTableExists, func(stmt *Statement) error {
		exists = true
		return nil
	}, map[string]interface{}{
		"$name": name,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (c *Connection) ReOpen() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.Conn.Close()
	if err != nil {
		return err
	}

	return c.openConn()
}

// pollReOpen polls the connection and reopens it after the given interval
// if no interval is given, it will default to 5 seconds
func (c *Connection) pollReOpen(interval ...time.Duration) {
	if len(interval) == 0 {
		interval = []time.Duration{time.Second * 5}
	}

	consecutiveFailures := 0

	for {
		time.Sleep(interval[0])
		if c.closed {
			break
		}
		if consecutiveFailures > 5 {
			log.Printf("failed to reopen sqlite connection 5 times in a row, giving up")
			break
		}

		err := c.ReOpen()
		if err != nil {
			consecutiveFailures++
			log.Printf("failed to reopen sqlite connection during poll: %s", err)
		}
	}
}
