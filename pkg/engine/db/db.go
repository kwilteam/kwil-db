/*
Package db provides a database abstraction layer for SQLite.
This database is used as the underlying data store for Kwil databases.

This package is essentially a wrapper around SQLite, with functionality to
handle the database schema and metadata.
*/

package db

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/pkg/sql"
)

type DB struct {
	Sqldb SqlDB

	// caches metadata from QueryUnsafe.
	// this is a really bad practice, but works.
	// essentially, we cache the metadata the first time it is retrieved, during schema
	// deployment.  This prevents the need from calling QueryUnsafe again
	metadataCache map[metadataType][]*metadata

	mu sync.RWMutex
}

func (d *DB) Close() error {
	return d.Sqldb.Close()
}

func (d *DB) Delete() error {
	return d.Sqldb.Delete()
}

func (d *DB) Prepare(ctx context.Context, query string) (*PreparedStatement, error) {
	tables, err := d.ListTables(ctx)
	if err != nil {
		return nil, err
	}

	analyzed, err := sqlanalyzer.ApplyRules(query, sqlanalyzer.AllRules, &sqlanalyzer.RuleMetadata{
		Tables: tables,
	})
	if err != nil {
		return nil, err
	}

	prepStmt, err := d.Sqldb.Prepare(analyzed.Statement())
	if err != nil {
		// TODO: add logs here
		fmt.Println(query)
		fmt.Println(analyzed.Statement())
		return nil, err
	}

	return &PreparedStatement{
		Statement: prepStmt,
		mutative:  analyzed.Mutative(),
	}, nil
}

func (d *DB) Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error) {
	return d.Sqldb.Query(ctx, stmt, args)
}

func (d *DB) Savepoint() (sql.Savepoint, error) {
	return d.Sqldb.Savepoint()
}

func NewDB(ctx context.Context, sqldb SqlDB) (*DB, error) {
	db := &DB{
		Sqldb:         sqldb,
		metadataCache: make(map[metadataType][]*metadata),
	}

	err := db.initMetadataTable(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (d *DB) CreateSession() (sql.Session, error) {
	return d.Sqldb.CreateSession()
}

func (d *DB) ApplyChangeset(changeset io.Reader) error {
	return d.Sqldb.ApplyChangeset(changeset)
}
