package executor

import (
	"context"
	"fmt"
	"kwil/kwil/repository"
	"kwil/x/graphql/hasura"
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

	// untrack tables
	tables, err := dao.ListTables(ctx, &repository.ListTablesParams{
		DbName:         database.Name,
		AccountAddress: database.Owner,
	})
	if err != nil {
		return fmt.Errorf("error listing tables: %d", err)
	}
	if len(tables) == 0 {
		return fmt.Errorf("database does not have any tables")
	}

	for _, table := range tables {
		err = s.hasura.UntrackTable(hasura.DefaultSource, schemaName, table)
		if err != nil {
			return fmt.Errorf("error untracking table %s: %w", table, err)
		}
	}

	// drop the database from the databases table
	err = dao.DropDatabase(ctx, &repository.DropDatabaseParams{
		DbName:         database.Name,
		AccountAddress: database.Owner,
	})
	if err != nil {
		return fmt.Errorf("error dropping database from database table: %d", err)
	}

	// drop the postgres schema
	_, err = tx.ExecContext(ctx, "DROP SCHEMA $1", schemaName)
	if err != nil {
		return fmt.Errorf("error dropping schema %s. error: %d", schemaName, err)
	}

	return tx.Commit()
}
