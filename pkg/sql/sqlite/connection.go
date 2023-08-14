package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite/functions"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/go-sqlite/sqlitex"
	"go.uber.org/zap"
)

type Connection struct {
	conn        *sqlite.Conn
	mu          lockable // mutex to protect the write connection, using an interface to allow for nil mutex
	log         log.Logger
	path        string
	readPool    *sqlitex.Pool
	poolSize    int
	flags       sqlite.OpenFlags
	isMemory    bool
	name        string
	attachedDBs map[string]string // maps the name to the file name
}

// OpenConn opens a connection to the database with the given name.
// It takes optional ConnectionOptions, which can be used to specify the path, logger, and other options.
func OpenConn(name string, opts ...ConnectionOption) (*Connection, error) {
	connection := &Connection{
		log:         log.NewNoOp(),
		mu:          &sync.Mutex{},
		path:        DefaultPath,
		name:        name,
		poolSize:    10,
		conn:        nil,
		readPool:    nil,
		isMemory:    false,
		flags:       sqlite.OpenWAL | sqlite.OpenURI,
		attachedDBs: make(map[string]string),
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

	err = connection.EnableForeignKey()
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign key: %w", err)
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
	return c.formatFilePath(c.name)
}

func (c *Connection) formatFilePath(fileName string) string {
	return fmt.Sprintf("%s%s.sqlite", c.path, fileName)
}

func (c *Connection) openConn() error {
	var err error
	c.conn, err = sqlite.OpenConn(c.getFilePath(), c.openFlags(false))
	if err != nil {
		return fmt.Errorf("failed to open readwrite connection: %w", err)
	}

	err = functions.Register(c.conn)
	if err != nil {
		return fmt.Errorf("failed to register custom functions: %w", err)
	}

	err = c.attachDBs(c.conn)
	if err != nil {
		return fmt.Errorf("failed to attach databases: %w", err)
	}

	err = c.initializeReadPool()
	if err != nil {
		return fmt.Errorf("failed to initialize read pool: %w", err)
	}

	return nil
}

// initializeReadPool initializes the read connection pool.
// it will ensure attached databases and sqlite extensions are registered.
func (c *Connection) initializeReadPool() error {
	var err error
	c.readPool, err = sqlitex.Open(c.getFilePath(), c.openFlags(true), c.poolSize)
	if err != nil {
		return fmt.Errorf("failed to create read connection pool: %w", err)
	}

	poolArray := make([]*sqlite.Conn, c.poolSize)

	for i := 0; i < c.poolSize; i++ {
		conn := c.readPool.Get(context.Background())
		if conn == nil {
			return fmt.Errorf("failed to get read connection from connection pool")
		}

		poolArray[i] = conn
	}

	for _, conn := range poolArray {
		err = functions.Register(conn)
		if err != nil {
			return fmt.Errorf("failed to register custom functions: %w", err)
		}

		err = c.attachDBs(conn)
		if err != nil {
			return fmt.Errorf("failed to attach databases: %w", err)
		}
	}

	for _, conn := range poolArray {
		c.readPool.Put(conn)
	}

	return nil
}

// attachDBs attaches the databases to the connection
func (c *Connection) attachDBs(conn *sqlite.Conn) error {
	for name, file := range c.attachedDBs {
		err := attachDB(conn, name, c.formatURI(file))
		if err != nil {
			return fmt.Errorf("failed to attach database: %w", err)
		}
	}

	return nil
}

func attachDB(c *sqlite.Conn, schemaName, uri string) error {
	return sqlitex.ExecuteTransient(c, fmt.Sprintf(sqlAttach, uri, schemaName), nil)
}

const sqlAttach = `ATTACH DATABASE '%s' AS %s;`

// formatURI formats the URI for the given file name.
// It includes a read only flag
func (c *Connection) formatURI(fileName string) string {
	return fmt.Sprintf("file:%s?mode=ro", c.formatFilePath(fileName))
}

func (c *Connection) mkPathDir() error {
	dir := filepath.Dir(c.path)
	return os.MkdirAll(dir, os.ModePerm)
}

// execute executes a statement on the write connection.
// this should really only be used for DDL statements or pragmas.
// dml statements should use prepared statements instead unless they are one-offs.
// this method is intentionally barebones to prevent misuse.
func (c *Connection) Execute(stmt string, args ...map[string]any) error {

	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.execute(stmt, args...)
}

// execute executes a one-off statement.  It does not use a mutex, unlike Execute.
func (c *Connection) execute(stmt string, args ...map[string]any) error {
	cleanedStmt := trimPadding(stmt)

	if len(args) == 0 {
		return sqlitex.ExecuteTransient(c.conn, cleanedStmt, nil)
	}

	for _, arg := range args {
		err := sqlitex.ExecuteTransient(c.conn, cleanedStmt, &sqlitex.ExecOptions{
			Named: arg,
		})
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return nil
}

func (c *Connection) Prepare(stmt string) (*Statement, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	innerStmt, trailingBytes, err := c.conn.PrepareTransient(trimPadding(stmt))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	if trailingBytes > 0 { // there should not be trailing bytes since we use trimPadding
		return nil, fmt.Errorf("trailing bytes after statement: %q", trailingBytes)
	}

	return newStatement(c, innerStmt), nil
}

// Close closes the connection.
// It takes an optional wait channel, which will be waited on until the connection is closed.
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.close()
}

// close closes the connection.
// It takes an optional wait channel, which will be waited on before the connection is closed.
func (c *Connection) close() error {
	if c.conn == nil { // if the connection is nil, it's already closed / been deleted
		return nil
	}

	if c.readPool != nil {
		err := c.readPool.Close()
		if err != nil {
			return fmt.Errorf("failed to close read connection pool: %w", err)
		}
	}

	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

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

		if stmt == nil {
			return nil
		}

		return stmt.stmt.Finalize()
	}

	innerStmt, trailingBytes, err := readConn.PrepareTransient(trimPadding(statement))
	if err != nil {
		return nil, deferFunc, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if trailingBytes > 0 {
		return nil, deferFunc, fmt.Errorf("trailing bytes after statement: %q", trailingBytes)
	}

	return c.newReadOnlyStatement(readConn, innerStmt), deferFunc, nil
}

// Query executes a read-only query against the database.
// It takes a QueryOpts struct, which can contain arguments, a function to manually bind parameters, a function
// to manually handle each result in between Step() calls, and a struct to store the results in.
// All of these are optional, and if not provided, the function will return an error
// TODO: rename this to BeginQuery
func (c *Connection) Query(ctx context.Context, statement string, options ...ExecOption) (*Results, error) {
	if c.readPool == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	results, err := c.query(ctx, statement, removeNilVals(options)...)
	if err != nil {
		c.log.Error("failed to execute query", zap.Error(err))
		return nil, err
	}

	return results, nil
}

func removeNilVals[T any](vals []T) []T {
	newVals := make([]T, 0)
	for _, val := range vals {
		if !reflect.ValueOf(&val).Elem().IsZero() {
			newVals = append(newVals, val)
		}
	}
	return newVals
}

// query executes a query and calls the resultFn for each row returned
// statementSetterFn is a function that is called to set the arguments for the statement
func (c *Connection) query(ctx context.Context, statement string, options ...ExecOption) (*Results, error) {
	stmt, deferFunc, err := c.prepareRead(ctx, statement)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error preparing read: %w", err), deferFunc())
	}

	results, err := stmt.execute(ctx, options...)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error executing read: %w", err), deferFunc())
	}

	results.addCloser(deferFunc)

	return results, nil
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

func (c *Connection) ListTables(ctx context.Context) ([]string, error) {
	tables := make([]string, 0)
	results, err := c.Query(ctx, sqlListTables)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer results.Finish()

	for {
		rowReturned, err := results.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next row: %w", err)
		}

		if !rowReturned {
			break
		}

		row := results.GetRecord()

		nameAny, ok := row["name"]
		if !ok {
			return nil, fmt.Errorf("failed to get name from row: %w", err)
		}

		name, ok := nameAny.(string)
		if !ok {
			return nil, fmt.Errorf("failed to cast name to string: %w", err)
		}

		tables = append(tables, name)
	}

	return tables, nil
}

func (c *Connection) TableExists(ctx context.Context, tableName string) (bool, error) {
	exists := false
	results, err := c.Query(ctx, sqlIfTableExists,
		WithNamedArgs(map[string]interface{}{
			"$name": tableName,
		}),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if table exists: %w", err)
	}
	defer results.Finish()

	for {
		rowReturned, err := results.Next()
		if err != nil {
			return false, fmt.Errorf("failed to get next row: %w", err)
		}

		if !rowReturned {
			break
		}

		exists = true
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

	err := c.close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

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

// ApplyChangeset applies a changeset to the database.
// It will either all succeed or all fail.
func (c *Connection) ApplyChangeset(r io.Reader) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.ApplyChangeset(r, nil, func(ct sqlite.ConflictType, ci *sqlite.ChangesetIterator) sqlite.ConflictAction {
		op, err := ci.Operation()
		if err != nil {
			fmt.Println("Error getting operation: ", err)
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

/*
for some fucking reason this will not work, we don't even need it yet so will come back when we do
the invertedChangset is null, however r is (AFAIK) a valid changeset
// InverseChangeset applies the inverse of the given changeset.
func (c *Connection) InverseChangeset(r *bytes.Buffer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	invertedChangeset := new(bytes.Buffer)
	err := sqlite.InvertChangeset(r, invertedChangeset)
	if err != nil {
		return fmt.Errorf("failed to invert changeset: %w", err)
	}

	if invertedChangeset.Len() == 0 {
		return fmt.Errorf("inverted changeset is empty")
	}

	return c.conn.ApplyInverseChangeset(invertedChangeset, nil, nil)
}
*/

type ResultSet struct {
	Rows    [][]any  `json:"rows"`
	Columns []string `json:"columns"`
	index   int
}

// Next increments the row index by 1 and returns true if there is another row in the result set
func (r *ResultSet) Next() bool {
	r.index++
	return r.index < len(r.Rows)
}

// GetColumn returns the value of the column at the given index
func (r *ResultSet) GetColumn(name string) any {
	for i, col := range r.Columns {
		if col == name {
			return r.Rows[r.index][i]
		}
	}
	return nil
}

func (r *ResultSet) GetRecord() map[string]any {
	record := make(map[string]any)
	for i, col := range r.Columns {
		record[col] = r.Rows[r.index][i]
	}
	return record
}

// Records will retrieve all records for the result set.
// It will not reset the row index.
func (r *ResultSet) Records() []map[string]any {
	records := make([]map[string]any, len(r.Rows))
	for i, row := range r.Rows {
		record := make(map[string]any)
		for j, col := range r.Columns {
			record[col] = row[j]
		}
		records[i] = record
	}

	return records
}

// Reset resets the row index to -1
func (r *ResultSet) Reset() {
	r.index = -1
}
