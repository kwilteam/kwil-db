package executor

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/internal/repository"
	"kwil/pkg/databases"
	"kwil/pkg/databases/clean"
	"kwil/pkg/databases/convert"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/spec"
	schemabuilder "kwil/pkg/databases/sql-builder/schema-builder"

	"go.uber.org/zap"
)

func (s *executor) DeployDatabase(ctx context.Context, database *databases.Database[*spec.KwilAny]) error {
	schemaName := database.GetSchemaName()

	// check if database exists
	if s.databaseExists(schemaName) {
		return fmt.Errorf(`database id "%s" already exists`, database.GetSchemaName())
	}

	// clean database
	clean.Clean(database)

	// validate database
	err := s.ValidateDatabase(database)
	if err != nil {
		return fmt.Errorf(`error on database "%s": %w`, database.GetSchemaName(), err)
	}

	// tx to be used to store database
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf(`error creating transaction when deploying database "%s": %w`, database.GetSchemaName(), err)
	}

	// manages the creation of the database
	creator := s.newDbCreator(ctx, database, tx)

	// store database
	err = creator.Store(ctx)
	if err != nil {
		return fmt.Errorf(`error storing database "%s": %w`, database.GetSchemaName(), err)
	}

	// commit tx
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(`error on database "%s": %w`, database.GetSchemaName(), err)
	}

	err = s.Track(database)
	if err != nil {
		s.log.Error("error tracking database", zap.String("database", database.GetSchemaName()), zap.Error(err))
	}

	// we add the database after tx commit since this only gets emptied when a DB is deleted or when the server is restarted
	dbInterface, err := executables.FromDatabase(database)
	if err != nil {
		return fmt.Errorf("failed to generate database interface: %w", err)
	}

	// add database to in-memory cache
	s.databases[database.GetSchemaName()] = dbInterface

	return nil
}

type dbCreator struct {
	database *databases.Database[*spec.KwilAny]
	dao      repository.Queries
	tx       *sql.Tx
}

// newDbCreator creates a new dbCreator for storing the given database
func (s *executor) newDbCreator(ctx context.Context, db *databases.Database[*spec.KwilAny], tx *sql.Tx) *dbCreator {
	dao := s.dao.WithTx(tx)
	return &dbCreator{
		database: db,
		dao:      dao,
		tx:       tx,
	}
}

// Store calls all of the internal store methods in the required order
func (d *dbCreator) Store(ctx context.Context) error {
	err := d.storeDatabase(ctx)
	if err != nil {
		return err
	}

	dbid, err := d.dao.GetDatabaseId(ctx, &databases.DatabaseIdentifier{
		Name:  d.database.Name,
		Owner: d.database.Owner,
	})
	if err != nil {
		return fmt.Errorf("error getting database id: %w", err)
	}

	err = d.storeTables(ctx, dbid)
	if err != nil {
		return err
	}

	err = d.storeQueries(ctx, dbid)
	if err != nil {
		return err
	}

	err = d.storeRoles(ctx, dbid)
	if err != nil {
		return err
	}

	err = d.storeIndexes(ctx, dbid)
	if err != nil {
		return err
	}

	return d.buildDatabase(ctx)
}

// creates the database in the database table
func (d *dbCreator) storeDatabase(ctx context.Context) error {
	return d.dao.CreateDatabase(ctx, &databases.DatabaseIdentifier{
		Name:  d.database.Name,
		Owner: d.database.Owner,
	})
}

// stores tables, columns, and attributes
func (d *dbCreator) storeTables(ctx context.Context, dbid int32) error {
	for _, table := range d.database.Tables {
		err := d.dao.CreateTable(ctx, dbid, table.Name)
		if err != nil {
			return fmt.Errorf("error storing table: %w", err)
		}

		tableId, err := d.dao.GetTableId(ctx, dbid, table.Name)
		if err != nil {
			return fmt.Errorf("error getting table id: %w", err)
		}

		// create columns
		for _, column := range table.Columns {
			err := d.dao.CreateColumn(ctx, tableId, column.Name, int32(column.Type))
			if err != nil {
				return fmt.Errorf("error storing column: %w", err)
			}

			var columnId int32
			if len(column.Attributes) > 0 {
				columnId, err = d.dao.GetColumnId(ctx, tableId, column.Name)
				if err != nil {
					return fmt.Errorf("error getting column id: %w", err)
				}
			}

			// create attributes
			for _, attribute := range column.Attributes {
				err = d.dao.CreateAttribute(ctx, columnId, int32(attribute.Type), attribute.Value)
				if err != nil {
					return fmt.Errorf("error storing attribute: %w", err)
				}
			}
		}
	}
	return nil
}

// stores queries.
// we don't have to do anything with the modifiers since they just get included in the query BLOB
func (d *dbCreator) storeQueries(ctx context.Context, dbid int32) error {
	for _, query := range d.database.SQLQueries {
		// convert query from kwilany to bytes and encode it with gob
		bts, err := convert.KwilAny.SQLQueryToBytes(query).EncodeGOB()
		if err != nil {
			return fmt.Errorf("error serializing query: %w", err)
		}

		err = d.dao.CreateQuery(ctx, query.Name, dbid, bts)
		if err != nil {
			return fmt.Errorf("error storing query: %w", err)
		}
	}

	return nil
}

// stores roles. must be called after createQueries
func (d *dbCreator) storeRoles(ctx context.Context, dbid int32) error {
	for _, role := range d.database.Roles {
		err := d.dao.CreateRole(ctx, dbid, role.Name, role.Default)
		if err != nil {
			return fmt.Errorf("error storing role: %w", err)
		}

		// create role permissions
		for _, permission := range role.Permissions {
			err = d.dao.ApplyPermissionToRole(ctx, dbid, role.Name, permission)
			if err != nil {
				return fmt.Errorf("error applying query to role: %w", err)
			}
		}
	}

	return nil
}

// stores indexes. must be called after createTables
func (d *dbCreator) storeIndexes(ctx context.Context, dbid int32) error {
	for _, index := range d.database.Indexes {
		// get table id
		tableId, err := d.dao.GetTableId(ctx, dbid, index.Table)
		if err != nil {
			return fmt.Errorf("error getting table id: %w", err)
		}

		err = d.dao.CreateIndex(ctx, tableId, index.Name, int32(index.Using), index.Columns)
		if err != nil {
			return fmt.Errorf("error storing index: %w", err)
		}
	}
	return nil
}

// buildDatatabase builds the database from the database table.  The schema name is the sha224 hash prepended with an x
func (d *dbCreator) buildDatabase(ctx context.Context) error {
	// create schema
	err := d.dao.CreateSchema(ctx, d.database.GetSchemaName())
	if err != nil {
		return fmt.Errorf("error creating database schema: %w", err)
	}

	// generate ddl
	ddl, err := schemabuilder.GenerateDDL(d.database)
	if err != nil {
		return fmt.Errorf("error generating ddl: %w", err)
	}

	// execute ddl
	_, err = d.tx.ExecContext(ctx, ddl)
	if err != nil {
		err2 := d.dao.DropSchema(ctx, d.database.GetSchemaName())
		if err2 != nil {
			return fmt.Errorf("error dropping schema on deployment failure: drop schema error: %w.  deployment error: %w", err2, err)
		}
		return fmt.Errorf("error executing ddl: %w", err)
	}

	return nil
}
