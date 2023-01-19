package dbretriever

import (
	"context"
	"fmt"
	"kwil/kwil/repository/gen"
	"kwil/x/types/databases"
	"strings"
)

// GetDatabase returns a database object for the given database identifier
// The database should be cleaned after it is retrieved
func (q *dbRetriever) GetDatabase(ctx context.Context, dbIdent *databases.DatabaseIdentifier) (*databases.Database, error) {
	db := &databases.Database{
		Name:  strings.ToLower(dbIdent.Name),
		Owner: strings.ToLower(dbIdent.Owner),
	}

	dbid, err := q.gen.GetDatabaseId(ctx, &gen.GetDatabaseIdParams{
		DbName:         db.Name,
		AccountAddress: db.Owner,
	})
	if err != nil {
		return nil, fmt.Errorf(`error getting database id for %s: %w`, dbIdent, err)
	}

	// get tables
	tables, err := q.GetTables(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting tables for %s: %w`, dbIdent, err)
	}

	// get queries
	queries, err := q.GetQueries(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting queries for %s: %w`, dbIdent, err)
	}

	// get roles
	roles, err := q.GetRoles(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting roles for %s: %w`, dbIdent, err)
	}

	// get indexes
	indexes, err := q.GetIndexes(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting indexes for %s: %w`, dbIdent, err)
	}

	db.Tables = tables
	db.SQLQueries = queries
	db.Roles = roles
	db.Indexes = indexes

	return db, nil

}

func (q *dbRetriever) ListDatabases(ctx context.Context) ([]*databases.DatabaseIdentifier, error) {
	res, err := q.gen.ListDatabases(ctx)
	if err != nil {
		return nil, err
	}

	dbs := make([]*databases.DatabaseIdentifier, len(res))
	for i, db := range res {
		dbs[i] = &databases.DatabaseIdentifier{
			Name:  db.DbName,
			Owner: db.AccountAddress,
		}
	}

	return dbs, nil
}

func (q *dbRetriever) ListDatabasesByOwner(ctx context.Context, owner string) ([]string, error) {
	return q.gen.ListDatabasesByOwner(ctx, owner)
}
