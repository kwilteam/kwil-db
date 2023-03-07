package driver

import (
	"fmt"
	"sync"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const (
	FilePathSuffix = ".sqlite"
	DefaultPath    = "~./.kwil/sqlite/"
)

type Connection struct {
	conn      *sqlite.Conn
	mu        *sync.Mutex
	DBID      string
	lock      LockType
	path      string
	readOnly  bool
	savepoint *savepoint
}

// OpenConn opens a connection to the database with the given ID/name.
func OpenConn(dbid string, opts ...ConnOpt) (*Connection, error) {
	connection := &Connection{
		DBID:     dbid,
		mu:       &sync.Mutex{},
		lock:     LOCK_TYPE_UNLOCKED,
		path:     DefaultPath,
		readOnly: false,
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

	if c.conn == nil {
		conn, err := sqlite.OpenConn(c.getFilePath(), flags)
		if err != nil {
			return err
		}
		c.conn = conn
		c.lock = LOCK_TYPED_SHARED
	}

	if c.savepoint == nil {
		c.savepoint = newSavepoint(c.conn)
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
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
		c.conn = nil
	}
	return nil
}

// Execute executes a statement
func (c *Connection) Execute(stmt string, args ...interface{}) error {
	return c.execute(stmt, &sqlitex.ExecOptions{
		Args: args,
	})
}

// ExecuteNamed executes a statement
func (c *Connection) ExecuteNamed(stmt string, args map[string]interface{}) error {
	return c.execute(stmt, &sqlitex.ExecOptions{
		Named: args,
	})
}

func (c *Connection) execute(stmt string, options *sqlitex.ExecOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.Writable() {
		return ErrNoWriteLock
	}

	err := sqlitex.Execute(c.conn, stmt, options)
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
		return stmt.SetMany(args)
	})
}

// query executes a query and calls the resultFn for each row returned
// statementSetterFn is a function that is called to set the arguments for the statement
func (c *Connection) query(statement string, resultFn ResultFn, statementSetterFn func(*Statement) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	stmt, err := c.Prepare(statement)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

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

		if err := resultFn(stmt); err != nil {
			return err
		}
	}

	return nil
}

// Prepare prepares a statement for execution
func (c *Connection) Prepare(statement string) (*Statement, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.Readable() {
		return nil, ErrConnectionClosed
	}

	sqliteStmt, err := c.conn.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return newStatement(sqliteStmt), nil
}

func (c *Connection) ExecuteStatement(stmt *Statement) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.Writable() {
		return ErrNoWriteLock
	}

	if err := stmt.Clear(); err != nil {
		return fmt.Errorf("failed to reset statement: %w", err)
	}

	_, err = stmt.step()
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}
