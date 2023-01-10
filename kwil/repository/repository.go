package repository

import (
	"context"
	"database/sql"
	dbretriever "kwil/kwil/repository/db_retriever"
	"kwil/kwil/repository/gen"
	"kwil/x/sqlx/sqlclient"
)

type Queries interface {
	DatabaseBuilder
	dbretriever.DatabaseRetriever
	ChainSyncer
	Accounter
	WithTx(tx *sql.Tx) Queries
}

type queries struct {
	db          *sqlclient.DB
	gen         *gen.Queries
	dbRetriever dbretriever.DatabaseRetriever
}

func New(db *sqlclient.DB) Queries {
	qrs := gen.New(db)

	dbRet := dbretriever.New(qrs)

	return &queries{
		db:          db,
		gen:         qrs,
		dbRetriever: dbRet,
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
	}, nil
}

func (q *queries) WithTx(tx *sql.Tx) Queries {
	return &queries{
		db:  q.db,
		gen: q.gen.WithTx(tx),
	}
}
