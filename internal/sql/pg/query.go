package pg

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/internal/sql/v2" // temporary v2 for refactoring

	"github.com/jackc/pgx/v5"
)

// NamedArgs is a query rewriter that can be used as one of the first arguments
// in the []any provided to a query function so that the named arguments are
// automatically used to rewrite the SQL statement from named (using @argname)
// syntax to positional ($1, $2, etc.). IMPORTANT: Note that the input statement
// requires named arguments to us "@" instead of "$" for the named arguments.
// Modify the SQL string as necessary to work with this rewriter.
type NamedArgs = pgx.NamedArgs

var _ pgx.QueryRewriter = NamedArgs{}

// QueryMode is a type recognized by the query methods when in one of the first
// arguments in the []any that causes the query to be executed in a certain
// mode. Presently this is used to change from the prepare/describe approaches
// to determining input argument type to a simpler mode that infers the argument
// types from the passed Go variable types, which is helpful for "in-line
// expressions" such as `SELECT $1;` that convey no information on their own
// about the type of the argument, resulting in an assumed type that may not
// match the type of provided Go variable (error in many cases).
type QueryMode = pgx.QueryExecMode

var (
	// QueryModeDefault uses a prepare-query request cycle to determine arg
	// types (OID) using postgres to describe the statement. This may not be
	// helpful for in-line expressions that reference no known table. There must
	// be an encode/decode plan available for the OID and the Go type.
	QueryModeDefault QueryMode = pgx.QueryExecModeCacheStatement
	// QueryModeExec infers the argument types from the Go variable type.
	QueryModeExec QueryMode = pgx.QueryExecModeExec
	// QueryModeSimple is like QueryModeExec, except that it uses the "simple"
	// postgresql wire protocol. Prefer QueryModeExec if argument type inference
	// based on the Go variables is required.
	QueryModeSimple QueryMode = pgx.QueryExecModeSimpleProtocol

	// NOTE: both QueryModeExec and QueryModeSimple can work with types
	// registered using pgtype.Map.RegisterDefaultPgType.
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
		// NOTE: if "unable to encode %v into OID %d in text format", it may
		// require pgx.QueryExecModeSimpleProtocol
		return nil, err
	}

	// res := rows.CommandTag() // RowsAffected, bool for Select etc.
	resSet := &sql.ResultSet{}
	for _, colInfo := range rows.FieldDescriptions() {
		// fmt.Println(colInfo.DataTypeOID, colInfo.DataTypeSize)

		// NOTE: if the column Name is "?column?", then colInfo.TableOID is
		// probably zero, meaning not a column of a table, e.g. the result of an
		// aggregate function, or just returning the a bound argument directly.
		// AND no AS was used.
		resSet.ReturnedColumns = append(resSet.ReturnedColumns, colInfo.Name)
	}

	resSet.Rows, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) ([]any, error) {
		return rows.Values()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	ctag := rows.CommandTag()
	resSet.Status = sql.CommandTag{
		Text:         ctag.String(),
		RowsAffected: ctag.RowsAffected(),
	}
	// if err != nil { fmt.Printf("**** query error\n\n%v\n\n%v\n", stmt, err) }
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
