package sqldb

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb/serialize"
	sqlitegenerator "github.com/kwilteam/kwil-db/pkg/engine/sqldb/sql-ddl-generator"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

const (
	metadataTableName      = "_metadata"
	sqlCreateMetadataTable = `CREATE TABLE IF NOT EXISTS ` + metadataTableName + ` (
		name TEXT NOT NULL,
		meta_type TEXT NOT NULL,
		version INTEGER NOT NULL,
		data BLOB NOT NULL,
		PRIMARY KEY (meta_type, name)
	) WITHOUT ROWID;`
	sqlGetMetadata   = `SELECT version, data, name FROM ` + metadataTableName + ` WHERE meta_type = $metatype;`
	sqlStoreMetadata = `INSERT INTO ` + metadataTableName + ` (name, meta_type, version, data) VALUES ($name, $metatype, $version, $data);`
)

func (s *SqliteStore) initTables() error {
	ctx := context.Background()

	exists, err := s.conn.TableExists(ctx, metadataTableName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	err = s.conn.Execute(sqlCreateMetadataTable)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteStore) CreateTable(ctx context.Context, table *dto.Table) error {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	err = s.deployTable(ctx, table)
	if err != nil {
		return err
	}

	err = s.storeTable(ctx, table)
	if err != nil {
		return err
	}

	return sp.Commit()
}

// deployTable deploys a table to the database.
// If the table already exists, it is not deployed, and false is returned.
// If the table does not exist, it is deployed, and true is returned.
func (s *SqliteStore) deployTable(ctx context.Context, table *dto.Table) error {
	exists, err := s.conn.TableExists(ctx, table.Name)
	if err != nil {
		return fmt.Errorf("failed to check if table %s exists: %w", table.Name, err)
	}

	if exists {
		return fmt.Errorf("table %s already exists", table.Name)
	}

	stmts, err := sqlitegenerator.GenerateDDL(table)
	if err != nil {
		return fmt.Errorf("failed to generate DDL for table %s: %w", table.Name, err)
	}

	for _, stmt := range stmts {
		err := s.conn.Execute(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute DDL for table %s: %w", table.Name, err)
		}
	}

	return nil
}

func (s *SqliteStore) StoreAction(ctx context.Context, action *dto.Action) error {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	err = s.storeAction(ctx, action)
	if err != nil {
		return err
	}

	return sp.Commit()
}

func (s *SqliteStore) StoreExtension(ctx context.Context, ext *dto.ExtensionInitialization) error {
	sp, err := s.conn.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	err = s.storeExtension(ctx, ext)
	if err != nil {
		return err
	}

	return sp.Commit()
}

func (s *SqliteStore) getMetadata(ctx context.Context, metaType serialize.TypeIdentifier) ([]*serialize.Serializable, error) {
	res := make([]*serialize.Serializable, 0)

	err := s.conn.Query(ctx, sqlGetMetadata,
		sqlite.WithNamedArgs(map[string]any{
			"$metatype": metaType,
		}),
		sqlite.WithResultFunc(func(stmt *sqlite.Statement) error {
			ser := &serialize.Serializable{
				Type: metaType,
			}

			ser.Data = stmt.GetBytes("data")
			ser.Version = stmt.GetInt64("version")
			ser.Name = stmt.GetText("name")

			res = append(res, ser)

			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *SqliteStore) ListTables(ctx context.Context) ([]*dto.Table, error) {
	sers, err := s.getMetadata(ctx, serialize.IdentifierTable)
	if err != nil {
		return nil, err
	}

	var tables []*dto.Table
	for _, ser := range sers {
		table, err := ser.Table()
		if err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *SqliteStore) ListActions(ctx context.Context) ([]*dto.Action, error) {
	sers, err := s.getMetadata(ctx, serialize.IdentifierAction)
	if err != nil {
		return nil, err
	}

	var actions []*dto.Action
	for _, ser := range sers {
		action, err := ser.Action()
		if err != nil {
			return nil, err
		}

		actions = append(actions, action)
	}

	return actions, nil
}

func (s *SqliteStore) GetExtensions(ctx context.Context) ([]*dto.ExtensionInitialization, error) {
	sers, err := s.getMetadata(ctx, serialize.IdentifierExtension)
	if err != nil {
		return nil, err
	}

	var extensions []*dto.ExtensionInitialization
	for _, ser := range sers {
		ext, err := ser.Extension()
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, ext)
	}

	return extensions, nil
}

func (s *SqliteStore) storeTable(ctx context.Context, table *dto.Table) error {
	ser, err := serialize.SerializeTable(table)
	if err != nil {
		return err
	}

	return s.storeMetadata(ctx, ser)
}

func (s *SqliteStore) storeAction(ctx context.Context, action *dto.Action) error {
	ser, err := serialize.SerializeAction(action)
	if err != nil {
		return err
	}

	return s.storeMetadata(ctx, ser)
}

func (s *SqliteStore) storeExtension(ctx context.Context, ext *dto.ExtensionInitialization) error {
	ser, err := serialize.SerializeExtension(ext)
	if err != nil {
		return err
	}

	return s.storeMetadata(ctx, ser)
}

func (s *SqliteStore) storeMetadata(ctx context.Context, ser *serialize.Serializable) error {
	exists, err := s.exists(ctx, ser.Name, ser.Type)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("metadata of type '%s' and name '%s' already exists", ser.Type, ser.Name)
	}

	stmt, err := s.getMetadataStmt()
	if err != nil {
		return err
	}

	err = stmt.Execute(
		sqlite.WithNamedArgs(map[string]any{
			"$name":     ser.Name,
			"$metatype": ser.Type,
			"$version":  ser.Version,
			"$data":     ser.Data,
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

// getMetadataStmt retrives the metadata table insert statement.
// It is cached on the SqliteDB struct.
func (s *SqliteStore) getMetadataStmt() (*sqlite.Statement, error) {
	if s.metadataStmt != nil {
		return s.metadataStmt, nil
	}

	stmt, err := s.conn.Prepare(sqlStoreMetadata)
	if err != nil {
		return nil, err
	}

	s.metadataStmt = stmt

	return stmt, nil
}

// exists checks if metadata of that name and type exists
func (s *SqliteStore) exists(ctx context.Context, name string, metaType serialize.TypeIdentifier) (bool, error) {
	var exists bool

	err := s.conn.Query(ctx, `SELECT EXISTS(SELECT 1 FROM `+metadataTableName+` WHERE name = $name AND meta_type = $metatype);`,
		sqlite.WithNamedArgs(map[string]any{
			"$name":     name,
			"$metatype": metaType,
		}),
		sqlite.WithResultFunc(func(stmt *sqlite.Statement) error {
			exists = stmt.GetBool("exists")
			return nil
		}),
	)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *SqliteStore) TableExists(ctx context.Context, table string) (bool, error) {
	return s.conn.TableExists(ctx, table)
}
