/*
Package db provides a database abstraction layer for SQLite.
This database is used as the underlying data store for Kwil databases.

This package is essentially a wrapper around SQLite, with functionality to
handle the database schema and metadata.
*/

package db

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser"
)

type DB struct {
	sqldb SqlDB
}

func (d *DB) Close() error {
	return d.sqldb.Close()
}

func (d *DB) Delete() error {
	return d.sqldb.Delete()
}

func (d *DB) Prepare(query string) (Statement, error) {
	ast, err := sqlparser.Parse(query)
	if err != nil {
		return nil, err
	}

	generatedSql, err := ast.ToSQL()
	if err != nil {
		return nil, err
	}

	return d.sqldb.Prepare(generatedSql)
}

func (d *DB) Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error) {
	return d.sqldb.Query(ctx, stmt, args)
}

func (d *DB) Savepoint() (Savepoint, error) {
	return d.sqldb.Savepoint()
}

func NewDB(ctx context.Context, sqldb SqlDB) (*DB, error) {
	db := &DB{
		sqldb: sqldb,
	}

	err := db.initMetadataTable(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (d *DB) CreateSession() (Session, error) {
	return d.sqldb.CreateSession()
}

func (d *DB) ApplyChangeset(changeset io.Reader) error {
	return d.sqldb.ApplyChangeset(changeset)
}
