package schema

import (
	"context"
	"fmt"
	"kwil/x/sqlx/schema_manager"
)

// Store will store the database schema in the database passed in.
// Make sure to call Validate() before calling this method.
func (db *Database) Store(ctx context.Context, manager schema_manager.Manager) error {

	// check if the database exists
	schemaName := db.SchemaName() // setting this to a variable to avoid calling it multiple times
	exists, err := manager.SchemaExists(ctx, schemaName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("database %s already exists", schemaName)
	}

	// create the database
	err = manager.NewDB(ctx, schemaName)
	if err != nil {
		return db.revert(ctx, manager, err)
	}

	// if we successfully create the database, we should store the DDL
	err = db.storeDDL(ctx, manager)
	if err != nil {
		return db.revert(ctx, manager, err)
	}

	// if we successfully store the DDL, we should store the metadata
	err = db.storeMetadata(ctx, manager)
	if err != nil {
		return db.revert(ctx, manager, err)
	}

	return nil
}

// revertDB will try to delete the database if there is an error.
func (db *Database) revert(ctx context.Context, manager schema_manager.Manager, err error) error {
	deleteErr := db.Delete(ctx, manager)
	if deleteErr != nil {
		return fmt.Errorf("failed to create database %s: %w. failed to delete database: %v", db.SchemaName(), err, deleteErr)
	}
	return fmt.Errorf("failed to create database %s: %w", db.SchemaName(), err)
}

// storeDDL generates and stores the structure of the user's database.
// This includes tables, indexes, foreign keys, etc.
func (db *Database) storeDDL(ctx context.Context, manager schema_manager.Manager) error {
	stmts, err := db.GenerateDDL()
	if err != nil {
		return err
	}

	for _, stmt := range stmts {
		_, err = manager.Client().ExecContext(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

// storeMetadata stores things like roles and queries in the database.
func (db *Database) storeMetadata(ctx context.Context, manager schema_manager.Manager) error {

	schemaName := db.SchemaName()

	// adding queries
	queries := db.Queries.GetAll()
	for name, query := range queries {
		executable, err := query.Prepare(db)
		if err != nil {
			return fmt.Errorf("failed to prepare query %s: %w", name, err)
		}

		execBytes, err := executable.Bytes()
		if err != nil {
			return fmt.Errorf("failed to get bytes of query %s: %w", name, err)
		}

		err = manager.AddQuery(ctx, schemaName, name, execBytes)
		if err != nil {
			return fmt.Errorf("failed to add query %s: %w", name, err)
		}
	}

	for name, role := range db.Roles {
		err := manager.AddRole(ctx, schemaName, name)
		if err != nil {
			return fmt.Errorf("failed to add role: %w", err)
		}
		for _, queryName := range role.Queries {
			err = manager.AddQueryPermission(ctx, schemaName, name, queryName)
			if err != nil {
				return fmt.Errorf("failed to add query permission: %w", err)
			}
		}
	}

	return nil
}

// Delete will delete the schema from the database passed in from the manager.
func (db *Database) Delete(ctx context.Context, manager schema_manager.Manager) error {
	return manager.DeleteDB(ctx, db.SchemaName())
}
