package pg

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/internal/sql/v2" // temporary v2 for refactoring

	"github.com/jackc/pgx/v5"
)

// These functions adapt the pgx query functions to Kwil's that return a
// *sql.ResultSet. Note that exec requires no wrapper, only to discard the first
// return.

type queryFun func(ctx context.Context, stmt string, args ...any) (pgx.Rows, error)

func query(ctx context.Context, q queryFun, stmt string, args ...any) (*sql.ResultSet, error) {
	rows, err := q(ctx, stmt, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	// res := rows.CommandTag() // RowsAffected, bool for Select etc.
	resSet := &sql.ResultSet{}
	for _, colInfo := range rows.FieldDescriptions() {
		// fmt.Println(colInfo.DataTypeOID, colInfo.DataTypeSize)
		resSet.ReturnedColumns = append(resSet.ReturnedColumns, colInfo.Name)
	}

	resSet.Rows, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) ([]any, error) {
		return rows.Values()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	return resSet, err
}

type txBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func queryTx(ctx context.Context, dbTx txBeginner, stmt string, args ...any) (*sql.ResultSet, error) {
	var resSet *sql.ResultSet
	err := pgx.BeginTxFunc(ctx, dbTx,
		pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
			IsoLevel:   pgx.RepeatableRead,
		},
		func(tx pgx.Tx) error {
			var err error
			resSet, err = query(ctx, tx.Query, stmt, args...)
			return err
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, sql.ErrNoRows
	}

	return resSet, err
}
