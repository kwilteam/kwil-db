package executor

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
)

func (s *executor) DropDatabase(ctx context.Context, database *databases.DatabaseIdentifier) error {
	schemaName := databases.GenerateSchemaName(database.Owner, database.Name)

	// create tx
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %d", err)
	}
	defer tx.Commit()
	dao := s.dao.WithTx(tx)

	/*
		dbid, err := dao.GetDatabaseId(ctx, &databases.DatabaseIdentifier{
			Name:  database.Name,
			Owner: database.Owner,
		})
		if err != nil {
			return fmt.Errorf("error getting database id: %w", err)
		}

		// untrack tables
		tables, err := dao.ListTables(ctx, dbid)
		if err != nil {
			return fmt.Errorf("error listing tables: %d", err)
		}
		if len(tables) == 0 {
			return fmt.Errorf("database does not have any tables")
		}

		for _, table := range tables {
			err = s.hasura.UntrackTable(hasura.DefaultSource, schemaName, table.TableName)
			if err != nil {
				return fmt.Errorf("error untracking table %s: %w", table.TableName, err)
			}
		}
	*/

	// drop the database from the databases table
	err = dao.DropDatabase(ctx, &databases.DatabaseIdentifier{
		Name:  database.Name,
		Owner: database.Owner,
	})
	if err != nil {
		return fmt.Errorf("error dropping database from database table: %d", err)
	}

	// check if schema exists
	exists, err := dao.SchemaExists(ctx, schemaName)
	if err != nil {
		return fmt.Errorf("error checking if schema exists: %d", err)
	}
	if !exists {
		return fmt.Errorf("database id %s does not exist", schemaName)
	}

	// drop the postgres schema
	err = dao.DropSchema(ctx, schemaName)
	if err != nil {
		return fmt.Errorf("error dropping schema %s. error: %d", schemaName, err)
	}

	// delete from the in-memory executables
	delete(s.databases, schemaName)

	return tx.Commit()
}
