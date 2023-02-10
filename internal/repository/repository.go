package repository

import (
	"context"
	"database/sql"
	dbretriever "kwil/internal/repository/db_retriever"
	"kwil/internal/repository/gen"
	"kwil/internal/repository/schema"
	"kwil/pkg/sql/sqlclient"
)

type Queries interface {
	DatabaseBuilder
	dbretriever.DatabaseRetriever
	ChainSyncer
	Accounter
	schema.SchemaManager
	WithTx(tx *sql.Tx) Queries
}

type queries struct {
	db          *sqlclient.DB
	gen         *gen.Queries
	dbRetriever dbretriever.DatabaseRetrieverTxer
	schema      schema.SchemaManager
}

func New(db *sqlclient.DB) Queries {
	qrs := gen.New(db)
	return &queries{
		db:          db,
		gen:         qrs,
		dbRetriever: dbretriever.New(qrs),
		schema:      schema.New(db),
	}
}

func Prepare(ctx context.Context, db *sqlclient.DB) (Queries, error) {
	prep, err := gen.Prepare(ctx, db)
	if err != nil {
		return nil, err
	}

	dbRet := dbretriever.New(prep)

	return &queries{
		db:          db,
		gen:         prep,
		dbRetriever: dbRet,
		schema:      schema.New(db),
	}, nil
}

func (q *queries) WithTx(tx *sql.Tx) Queries {
	return &queries{
		db:          q.db,
		gen:         q.gen.WithTx(tx),
		dbRetriever: q.dbRetriever.WithTx(tx),
		schema:      q.schema,
	}
}
