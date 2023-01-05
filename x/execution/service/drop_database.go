package service

import (
	"context"
	"fmt"
	"kwil/x/execution/dto"
	"kwil/x/execution/repository"
)

func (s *executionService) DropDatabase(ctx context.Context, database *dto.DatabaseIdentifier) error {
	// create tx
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %d", err)
	}
	defer tx.Commit()
	dao := s.dao.WithTx(tx)

	// drop the database from the databases table
	err = dao.DropDatabase(ctx, &repository.DropDatabaseParams{
		DbName:  database.Name,
		DbOwner: database.Owner,
	})
	if err != nil {
		return fmt.Errorf("error dropping database from database table: %d", err)
	}

	// drop the postgres schema
	dbid := dto.GenerateSchemaName(database.Owner, database.Name)
	_, err = tx.ExecContext(ctx, "DROP SCHEMA $1", dbid)
	if err != nil {
		return fmt.Errorf("error dropping schema %s. error: %d", dbid, err)
	}

	return tx.Commit()
}
