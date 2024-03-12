package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/common/sql"

	"github.com/jackc/pgx/v5"
)

func queryImpliedArgTypes(ctx context.Context, conn *pgx.Conn, stmt string, args ...any) (pgx.Rows, error) {
	// To support named args with our "implied arg types" query mode, we need to
	// do the pop off any rewriter and rewrite the query and the args list
	// before attempting to determine the OID of each of the args.
	var queryRewriter pgx.QueryRewriter
optionLoop:
	for len(args) > 0 {
		switch arg := args[0].(type) {
		case QueryMode:
			return nil, fmt.Errorf("extra query mode specified: %v", arg)
		case pgx.QueryRewriter:
			queryRewriter = arg
			args = args[1:]
		default:
			break optionLoop
		}
	}

	if queryRewriter != nil {
		var err error
		stmt, args, err = queryRewriter.RewriteQuery(ctx, conn, stmt, args)
		if err != nil {
			return nil, fmt.Errorf("rewrite query failed: %w", err)
		}
	}

	tyMap := conn.TypeMap()
	var oids []uint32
	for _, arg := range args {
		ty, ok := tyMap.TypeForValue(arg)
		if !ok {
			return nil, errors.New("type unknown")
		}
		oids = append(oids, ty.OID)
	}

	pgConn := conn.PgConn()
	// unnamed prepare to get statement description with asserted OIDs
	sd, err := pgConn.Prepare(ctx, "", stmt, oids)
	if err != nil {
		return nil, err
	}

	// set the OIDs from args using the extended query builder
	var eqb pgx.ExtendedQueryBuilder
	err = eqb.Build(tyMap, sd, args)
	if err != nil {
		return nil, err
	}

	rdr := pgConn.ExecParams(ctx, stmt, eqb.ParamValues, sd.ParamOIDs, eqb.ParamFormats, eqb.ResultFormats)
	return pgx.RowsFromResultReader(tyMap, rdr), nil
}

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

const (
	// QueryModeDefault uses a prepare-query request cycle to determine arg
	// types (OID) using postgres to describe the statement. This may not be
	// helpful for in-line expressions that reference no known table. There must
	// be an encode/decode plan available for the OID and the Go type.
	QueryModeDefault QueryMode = pgx.QueryExecModeCacheStatement
	// QueryModeExec still uses the extended protocol, but does not ask
	// postgresql to describe the statement to determine types.
	QueryModeExec QueryMode = pgx.QueryExecModeExec
	// QueryModeSimple is like QueryModeExec, except that it uses the "simple"
	// postgresql wire protocol. Prefer QueryModeExec if argument type inference
	// based on the Go variables is required since this forces everything into text.
	QueryModeSimple QueryMode = pgx.QueryExecModeSimpleProtocol

	// NOTE: both QueryModeExec and QueryModeSimple can work with types
	// registered using pgtype.Map.RegisterDefaultPgType.

	// we claim the upper bits for our custom modes below. Could also use a
	// different type (not QueryMode) if this seems risky, but I like them grouped

	// QueryModeInferredArgTypes runs the query in a special execution mode that
	// is like QueryExecModeExec except that it infers the argument OIDs from
	// the Go argument values and asserts those types in preparing the
	// statement, which is necessary for our in-line expressions. It is
	// incompatible with other special arguments like NamedArgs.
	QueryModeInferredArgTypes QueryMode = 1 << 16
)

// These functions adapt the pgx query functions to Kwil's that return a
// *sql.ResultSet. Note that exec requires no wrapper, only to discard the first
// return.

// type queryFun func(ctx context.Context, stmt string, args ...any) (pgx.Rows, error)

func mustInferArgs(args []any) bool {
	if len(args) > 0 {
		mode, ok := args[0].(QueryMode)
		if ok {
			return mode == QueryModeInferredArgTypes
		}
	}
	return false
}

// connQueryer is satisfied by a pgx.Tx, and is used to either Query or access
// the underlying Conn. This might seem silly since *pgx.Conn itself has a query
// method, but pgx.Tx's Query has a little extra logic to ensure the transaction
// is active.
type connQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Conn() *pgx.Conn
}

var _ connQueryer = (*cqWrapper)(nil)

// cqWrapper implements connQueryer from a *pgx.Conn (as a pgx.Tx does).
// This looks
type cqWrapper struct {
	c *pgx.Conn
}

func (cq *cqWrapper) Conn() *pgx.Conn {
	return cq.c
}

func (cq *cqWrapper) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return cq.c.Query(ctx, sql, args...)
}

func query(ctx context.Context, cq connQueryer, stmt string, args ...any) (*sql.ResultSet, error) {
	q := cq.Query
	if mustInferArgs(args) {
		// return nil, errors.New("cannot use QueryModeInferredArgTypes with query")
		args = args[1:] // args[0] was QueryModeInferredArgTypes
		q = func(ctx context.Context, stmt string, args ...any) (pgx.Rows, error) {
			return queryImpliedArgTypes(ctx, cq.Conn(), stmt, args...)
		}
	}

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
		resSet.Columns = append(resSet.Columns, colInfo.Name)
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
			resSet, err = query(ctx, tx, stmt, args...)
			return err
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, sql.ErrNoRows
	}

	return resSet, err
}
