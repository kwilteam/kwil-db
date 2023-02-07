package executor

import (
	"context"
	"fmt"
	"kwil/pkg/execution/executables"
)

// load executables will load all the executables from every database
func (e *executor) loadExecutables(ctx context.Context) error {
	dbs, err := e.dao.ListDatabases(ctx)
	if err != nil {
		return fmt.Errorf(`error listing databases: %w`, err)
	}

	for _, db := range dbs {
		db, err := e.dao.GetDatabase(ctx, db)
		if err != nil {
			return fmt.Errorf(`error getting database %v: %w`, db, err)
		}

		// prepare
		dbInterface, err := executables.FromDatabase(db)
		if err != nil {
			return fmt.Errorf(`error preparing database executables %v: %w`, db, err)
		}

		// add to map
		e.databases[db.GetSchemaName()] = dbInterface
	}

	return nil
}
