package sqlite

import (
	"fmt"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite/functions"
)

// OpenReadOnlyMemory opens a read-only in-memory sqlite database.
func OpenReadOnlyMemory() (*MemoryConnection, error) {
	conn, err := sqlite.OpenConn(":memory:", sqlite.OpenMemory|sqlite.OpenReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to open readwrite connection: %w", err)
	}

	err = functions.Register(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register custom functions: %w", err)
	}

	return &MemoryConnection{conn: conn}, nil
}

type MemoryConnection struct {
	conn *sqlite.Conn
}

// Close closes the connection.
func (c *MemoryConnection) Close() error {
	return c.conn.Close()
}

// Query executes the given query with the given arguments.
func (c *MemoryConnection) Query(stmt string, args map[string]any) ([]map[string]any, error) {
	prepared, trailing, err := c.conn.PrepareTransient(trimPadding(stmt))
	if err != nil {
		return nil, err
	}
	if trailing > 0 {
		return nil, fmt.Errorf("trailing bytes after query: %d", trailing)
	}
	defer prepared.Finalize()

	err = setMany(prepared, args)
	if err != nil {
		return nil, err
	}

	// results holds the results of the query.
	results := make([]map[string]any, 0)

	// these are used to detect metadata about the query result.
	firstIter := true
	var columnNames []string
	var columnTypes []DataType
	for {
		hasRow, err := prepared.Step()
		if err != nil {
			return nil, err
		}

		if !hasRow {
			break
		}

		if firstIter {
			firstIter = false
			columnNames = determineColumnNames(prepared)
			columnTypes = determineColumnTypes(prepared)
		}

		results = append(results, getRecord(prepared, columnNames, columnTypes))
	}

	return results, nil
}
