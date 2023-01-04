package service

import (
	"context"
	"fmt"
	"kwil/x/execution/repository"
	"kwil/x/execution/utils"
)

func (s *executionService) DropDatabase(ctx context.Context, owner, name string) error {
	// create tx
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %d", err)
	}
	defer tx.Commit()
	dao := s.dao.WithTx(tx)

	// drop the database from the databases table
	err = dao.DropDatabase(ctx, &repository.DropDatabaseParams{
		DbName:  name,
		DbOwner: owner,
	})
	if err != nil {
		return fmt.Errorf("error dropping database from database table: %d", err)
	}

	// drop the postgres schema
	dbid := utils.GenerateSchemaName(owner, name)
	_, err = tx.ExecContext(ctx, "DROP SCHEMA $1", dbid)
	if err != nil {
		return fmt.Errorf("error dropping schema %s. error: %d", dbid, err)
	}

	return tx.Commit()
}
