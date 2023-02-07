package executor

import (
	"context"
	"fmt"
	"kwil/internal/pkg/graphql/hasura"
	"kwil/pkg/types/data_types/any_type"
	"kwil/pkg/types/databases"
)

// Track tracks the database in hasura and the database table in the database.
func (s *executor) Track(db *databases.Database[anytype.KwilAny]) error {
	schemaName := db.GetSchemaName()
	for _, table := range db.Tables {
		// track tables
		err := s.hasura.TrackTable(hasura.DefaultSource, schemaName, table.Name)
		if err != nil {
			return fmt.Errorf(`error tracking tables in Graphql on database "%s": %w`, db.GetSchemaName(), err)
		}
	}

	return nil
}

// Untrack untracks the database in hasura and the database table in the database.
func (s *executor) Untrack(ctx context.Context, name, owner string) error {

	dbid, err := s.dao.GetDatabaseId(ctx, &databases.DatabaseIdentifier{
		Name:  name,
		Owner: owner,
	})
	if err != nil {
		return fmt.Errorf("error getting database id: %w", err)
	}

	// untrack tables
	tables, err := s.dao.ListTables(ctx, dbid)
	if err != nil {
		return fmt.Errorf("error listing tables: %d", err)
	}
	if len(tables) == 0 {
		return fmt.Errorf("database does not have any tables")
	}

	schemaName := databases.GenerateSchemaName(owner, name)

	for _, table := range tables {
		err = s.hasura.UntrackTable(hasura.DefaultSource, schemaName, table.TableName)
		if err != nil {
			return fmt.Errorf("error untracking table %s: %w", table.TableName, err)
		}
	}

	return nil
}
