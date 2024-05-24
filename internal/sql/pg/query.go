package pg

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

	// convert all types to types registered in pgx's type map
	args, oids, err := encodeToPGType(args...)
	if err != nil {
		return nil, fmt.Errorf("encode to pg type failed: %w", err)
	}

	pgConn := conn.PgConn()
	// unnamed prepare to get statement description with asserted OIDs
	sd, err := pgConn.Prepare(ctx, "", stmt, oids)
	if err != nil {
		return nil, err
	}

	tyMap := conn.TypeMap()

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
	// QueryModeDescribeExec also uses a prepared statement, but it is unnamed
	// and not cached, and thus avoids the need to retry a query if the table
	// definitions are modified concurrently. Requires two round trips.
	QueryModeDescribeExec QueryMode = pgx.QueryExecModeDescribeExec
	// QueryModeExec still uses the extended protocol, but does not ask
	// postgresql to describe the statement to determine parameters types.
	// Instead, the types are determined from the Go variables.
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
	// is like QueryModeExec except that it infers the argument OIDs from the Go
	// argument values AND asserts those types in preparing the statement, which
	// is necessary for our in-line expressions. QueryModeExec does not use
	// Parse/Describe messages; this mode does while asserting the param types.
	// It is like a hybrid between QueryModeDescribeExec and QueryModeExec. It
	// is incompatible with other special arguments like NamedArgs.
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
		pgxVals, err := row.Values()
		if err != nil {
			return nil, err
		}
		return decodeFromPGType(pgxVals...)
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

// oidArrMap maps oids to their corresponding array oids.
// It only includes types that we care about in Kwil.
var oidArrMap = map[int]int{
	pgtype.BoolOID:    pgtype.BoolArrayOID,
	pgtype.ByteaOID:   pgtype.ByteaArrayOID,
	pgtype.Int8OID:    pgtype.Int8ArrayOID,
	pgtype.TextOID:    pgtype.TextArrayOID,
	pgtype.UUIDOID:    pgtype.UUIDArrayOID,
	pgtype.NumericOID: pgtype.NumericArrayOID,
}

// encodeToPGType encodes several Go types to their corresponding pgx types.
// It is capable of detecting special Kwil types and encoding them to their
// corresponding pgx types. It is only used if using inferred argument types.
// If not using inferred argument types, pgx will rely on the Valuer interface
// to encode the Go types to their corresponding pgx types.
// It also returns the pgx type OIDs for each value.
func encodeToPGType(vals ...any) ([]any, []uint32, error) {
	// encodeScalar is a helper function that converts a single value to a pgx.
	encodeScalar := func(v any) (any, int, error) {
		switch v := v.(type) {
		case nil:
			return nil, pgtype.TextOID, nil
		case bool:
			return v, pgtype.BoolOID, nil
		case int, int8, int16, int32, int64:
			return v, pgtype.Int8OID, nil
		case string:
			return v, pgtype.TextOID, nil
		case []byte:
			return v, pgtype.ByteaOID, nil
		case *types.UUID:
			return pgtype.UUID{Bytes: [16]byte(v.Bytes()), Valid: true}, pgtype.UUIDOID, nil
		case types.UUID:
			return pgtype.UUID{Bytes: [16]byte(v.Bytes()), Valid: true}, pgtype.UUIDOID, nil
		case [16]byte:
			return pgtype.UUID{Bytes: v, Valid: true}, pgtype.UUIDOID, nil
		case decimal.Decimal:
			return pgtype.Numeric{
				Int:   v.BigInt(),
				Exp:   v.Exp(),
				Valid: true,
			}, pgtype.NumericOID, nil
		case *decimal.Decimal:
			return pgtype.Numeric{
				Int:   v.BigInt(),
				Exp:   v.Exp(),
				Valid: true,
			}, pgtype.NumericOID, nil
		case types.Uint256:
			return pgtype.Numeric{
				Int:   v.ToBig(),
				Exp:   0,
				Valid: true,
			}, pgtype.NumericOID, nil
		case *types.Uint256:
			return pgtype.Numeric{
				Int:   v.ToBig(),
				Exp:   0,
				Valid: true,
			}, pgtype.NumericOID, nil
		}

		return nil, 0, fmt.Errorf("unsupported type: %T", v)
	}

	// we convert all types to postgres's type. If the underlying type is an
	// array, we will set it as that so that pgx can handle it properly.
	// The one exception is []byte, which is handled by pgx as a bytea.
	pgxVals := make([]any, len(vals))
	oids := make([]uint32, len(vals))
	for i, val := range vals {
		// if nil, we just set it to text.
		if val == nil {
			pgxVals[i] = nil
			oids[i] = pgtype.TextOID
			continue
		}

		dt := reflect.TypeOf(vals[i])
		if (dt.Kind() == reflect.Slice || dt.Kind() == reflect.Array) && dt.Elem().Kind() != reflect.Uint8 {
			valueOf := reflect.ValueOf(val)
			arr := make([]any, valueOf.Len())
			var oid int
			var err error
			for j := 0; j < valueOf.Len(); j++ {
				arr[j], oid, err = encodeScalar(valueOf.Index(j).Interface())
				if err != nil {
					return nil, nil, err
				}
			}
			pgxVals[i] = arr

			// the oid can be 0 if the array is empty. In that case, we just
			// set it to text array, since we cannot infer it from an empty array.
			if oid == 0 {
				oids[i] = pgtype.TextArrayOID
			} else {
				oids[i] = uint32(oidArrMap[oid])
			}
		} else {
			var err error
			var oid int
			pgxVals[i], oid, err = encodeScalar(val)
			if err != nil {
				return nil, nil, err
			}
			oids[i] = uint32(oid)
		}
	}

	return pgxVals, oids, nil
}

// decodeFromPGType decodes several pgx types to their corresponding Go types.
// It is capable of detecting special Kwil types and decoding them to their
// corresponding Go types.
func decodeFromPGType(vals ...any) ([]any, error) {
	decodeScalar := func(v any) (any, error) {
		switch v := v.(type) {
		default:
			return v, nil
		case pgtype.UUID:
			u := types.UUID(v.Bytes)
			return &u, nil
		case [16]byte:
			u := types.UUID(v)
			return &u, nil
		case pgtype.Numeric:
			if v.NaN {
				return "NaN", nil
			}

			// if we give postgres a number 5000, it will return it as 5 with exponent 3.
			// Since kwil's decimal semantics do not allow negative scale, we need to multiply
			// the number by 10^exp to get the correct value.
			if v.Exp > 0 {
				z := new(big.Int)
				z.Exp(big.NewInt(10), big.NewInt(int64(v.Exp)), nil)
				z.Mul(z, v.Int)
				v.Int = z
				v.Exp = 0
			}

			// there is a bit of an edge case here, where uint256 can be returned.
			// since most results simply get returned to the user via JSON, it doesn't
			// matter too much right now, so we'll leave it as-is.
			return decimal.NewFromBigInt(v.Int, v.Exp)
		}
	}

	goVals := make([]any, len(vals))
	for i, val := range vals {
		if val == nil {
			goVals[i] = nil
			continue
		}

		dt := reflect.TypeOf(vals[i])
		if (dt.Kind() == reflect.Slice || dt.Kind() == reflect.Array) && dt.Elem().Kind() != reflect.Uint8 {
			// we need to reflect the first type of the slice to determine what type the slice is.
			// if empty, we return the slice as is.
			valueOf := reflect.ValueOf(val)

			length := valueOf.Len()
			if length == 0 {
				goVals[i] = val
				continue
			}

			arr := make([]any, length)
			for j := 0; j < length; j++ {
				var err error
				arr[j], err = decodeScalar(valueOf.Index(j).Interface())
				if err != nil {
					return nil, err
				}
			}

			goVals[i] = arr
		} else {
			var err error
			goVals[i], err = decodeScalar(val)
			if err != nil {
				return nil, err
			}
		}
	}

	return goVals, nil
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
