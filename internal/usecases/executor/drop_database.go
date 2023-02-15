package executor

import (
	"context"
	"fmt"
	"kwil/pkg/databases"

	"go.uber.org/zap"
)

func (s *executor) DropDatabase(ctx context.Context, database *databases.DatabaseIdentifier) error {
	schemaName := databases.GenerateSchemaId(database.Owner, database.Name)

	// create tx
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %d", err)
	}
	defer tx.Commit()
	dao := s.dao.WithTx(tx)

	err = s.Untrack(ctx, database.Name, database.Owner)
	if err != nil {
		s.log.Error("error untracking database", zap.String("database", database.Name), zap.String("error", err.Error()))
	}

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
