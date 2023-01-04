package service

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/x/execution/clean"
	"kwil/x/execution/dto"
	"kwil/x/execution/executables"
	"kwil/x/execution/repository"
	schemabuilder "kwil/x/execution/sql-builder/schema-builder"
	"kwil/x/sqlx/sqlclient"

	"github.com/cstockton/go-conv"
)

func (s *executionService) DeployDatabase(ctx context.Context, database *dto.Database) error {
	// check if database exists
	if s.databaseExists(database.GetSchemaName()) {
		return fmt.Errorf(`database "%s" already exists`, database.GetSchemaName())
	}

	// clean database
	clean.CleanDatabase(database)

	// validate database
	err := s.ValidateDatabase(database)
	if err != nil {
		return fmt.Errorf(`error on database "%s": %w`, database.GetSchemaName(), err)
	}

	// tx to be used to store database
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// manages the creation of the database
	creator, err := s.newDbCreator(ctx, database, tx)
	if err != nil {
		return err
	}

	// store database
	err = creator.Store(ctx)
	if err != nil {
		return err
	}

	// commit tx
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(`error on database "%s": %w`, database.GetSchemaName(), err)
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
	database *dto.Database
	db       *sqlclient.DB
	dao      *repository.Queries
	tx       *sql.Tx
}

// newDbCreator creates a new dbCreator for storing the given database
func (s *executionService) newDbCreator(ctx context.Context, db *dto.Database, tx *sql.Tx) (*dbCreator, error) {
	dao := s.dao.WithTx(tx)
	return &dbCreator{
		database: db,
		dao:      dao,
		tx:       tx,
	}, nil
}

// Store calls all of the internal store methods in the required order
func (d *dbCreator) Store(ctx context.Context) error {
	err := d.storeDatabase(ctx)
	if err != nil {
		return err
	}

	err = d.storeTables(ctx)
	if err != nil {
		return err
	}

	err = d.storeQueries(ctx)
	if err != nil {
		return err
	}

	err = d.storeRoles(ctx)
	if err != nil {
		return err
	}

	err = d.storeIndexes(ctx)
	if err != nil {
		return err
	}

	return d.buildDatabase(ctx)
}

// creates the database in the database table
func (d *dbCreator) storeDatabase(ctx context.Context) error {
	return d.dao.CreateDatabase(ctx, &repository.CreateDatabaseParams{
		DbName:  d.database.Name,
		DbOwner: d.database.Owner,
	})
}

// stores tables, columns, and attributes
func (d *dbCreator) storeTables(ctx context.Context) error {
	for _, table := range d.database.Tables {
		err := d.dao.CreateTable(ctx, &repository.CreateTableParams{
			TableName: table.Name,
			DbName:    d.database.Name,
		})
		if err != nil {
			return fmt.Errorf("error storing table: %w", err)
		}

		// create columns
		for _, column := range table.Columns {
			err := d.dao.CreateColumn(ctx, &repository.CreateColumnParams{
				ColumnName: column.Name,
				TableName:  table.Name,
				ColumnType: int32(column.Type),
			})
			if err != nil {
				return fmt.Errorf("error storing column: %w", err)
			}

			// create attributes
			for _, attribute := range column.Attributes {
				stringValue, err := conv.String(attribute.Value)
				if err != nil {
					return fmt.Errorf("error converting attribute value to string: %w", err)
				}

				err = d.dao.CreateAttribute(ctx, &repository.CreateAttributeParams{
					ColumnName:     column.Name,
					AttributeType:  int32(attribute.Type),
					AttributeValue: []byte(stringValue),
				})
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
func (d *dbCreator) storeQueries(ctx context.Context) error {
	for _, query := range d.database.SQLQueries {
		bts, err := query.EncodeGOB()
		if err != nil {
			return fmt.Errorf("error serializing query: %w", err)
		}

		err = d.dao.CreateQuery(ctx, &repository.CreateQueryParams{
			QueryName: query.Name,
			Query:     bts,
			TableName: query.Table,
		})
		if err != nil {
			return fmt.Errorf("error storing query: %w", err)
		}
	}

	return nil
}

// stores roles. must be called after createQueries
func (d *dbCreator) storeRoles(ctx context.Context) error {
	for _, role := range d.database.Roles {
		err := d.dao.CreateRole(ctx, &repository.CreateRoleParams{
			RoleName:  role.Name,
			DbName:    d.database.Name,
			IsDefault: role.Default,
		})
		if err != nil {
			return fmt.Errorf("error storing role: %w", err)
		}

		// create role permissions
		for _, permission := range role.Permissions {
			err = d.dao.RoleApplyQuery(ctx, &repository.RoleApplyQueryParams{
				RoleName:  role.Name,
				QueryName: permission,
			})
			if err != nil {
				return fmt.Errorf("error applying query to role: %w", err)
			}
		}
	}

	return nil
}

// stores indexes. must be called after createTables
func (d *dbCreator) storeIndexes(ctx context.Context) error {
	for _, index := range d.database.Indexes {
		err := d.dao.CreateIndex(ctx, &repository.CreateIndexParams{
			IndexName: index.Name,
			TableName: index.Table,
			IndexType: int32(index.Using),
			Columns:   index.Columns,
		})
		if err != nil {
			return fmt.Errorf("error storing index: %w", err)
		}
	}
	return nil
}

// buildDatatabase builds the database from the database table.  The schema name is the sha224 hash prepended with an x
func (d *dbCreator) buildDatabase(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+d.database.Name)
	if err != nil {
		return fmt.Errorf("error creating database schema: %w", err)
	}

	// generate ddl
	ddl, err := schemabuilder.GenerateDDL(d.database)
	if err != nil {
		return fmt.Errorf("error generating ddl: %w", err)
	}

	// execute ddl
	_, err = d.db.ExecContext(ctx, ddl)
	if err != nil {
		return fmt.Errorf("error executing ddl: %w", err)
	}

	return nil
}
