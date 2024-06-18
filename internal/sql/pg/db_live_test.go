// go:build pglive

package pg

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/require"
	// "github.com/kwilteam/kwil-db/internal/conv"
)

func TestMain(m *testing.M) {
	// UseLogger(log.NewStdOut(log.InfoLevel))
	m.Run()
}

const (
	pingStmt = `-- ping`
)

var (
	cfg = &DBConfig{
		PoolConfig: PoolConfig{
			ConnConfig: ConnConfig{
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
		SchemaFilter: func(s string) bool {
			return strings.Contains(s, DefaultSchemaFilterPrefix)
		},
	}
)

// TestRollbackPreparedTxns tests the rollbackPreparedTxns in the following
// cases:
//
//  1. when there are no prepared transactions
//  2. when we create one and roll it back from the same connection
//  3. when we create one, disconnect, and make a fresh connection to rollback
//
// The final case is expected in crash recovery.
func TestRollbackPreparedTxns(t *testing.T) {
	ctx := context.Background()
	connStr := connString(cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DBName, false)
	cfg, _ := pgx.ParseConfig(connStr)
	warns := make(chan *pgconn.Notice, 1)
	var expectMessage string
	cfg.OnNotice = func(pc *pgconn.PgConn, n *pgconn.Notice) {
		if expectMessage != "" && strings.Contains(strings.ToLower(n.Message), expectMessage) {
			warns <- n
		}
	}
	cfg.OnPgError = func(_ *pgconn.PgConn, n *pgconn.PgError) bool { // for test debugging
		t.Logf("%v [%v]: %v / %v", n.Severity, n.Code, n.Message, n.Detail)
		return !strings.EqualFold(n.Severity, "FATAL")
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(ctx); err != nil {
			t.Error(err)
		}
	})

	// Run rollback with no prepared txns.
	_, err = rollbackPreparedTxns(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Exec(ctx, `create table if not exists prepared_test (x int8);`)
	if err != nil {
		t.Fatal(err)
	}

	// Make a prepared transaction
	var tx pgx.Tx

	readyPreparedTx := func() {
		tx, err = conn.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO prepared_test (x) VALUES (1);`); err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, `PREPARE TRANSACTION 'asdf1234';`)
		if err != nil {
			t.Fatal(err)
		}
	}

	readyPreparedTx()

	// test rollback from same connection.
	closed, err := rollbackPreparedTxns(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	if closed != 1 {
		t.Errorf("rolled back %d, wanted %d", closed, 1)
	}

	// the transaction is now over, Commit emits a warn notice, but no error returns
	expectMessage = "there is no transaction in progress"
	err = tx.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-warns:
	case <-time.After(2 * time.Second):
		t.Error("no warning received")
	}

	// test rollback from new connection
	readyPreparedTx()

	err = conn.Close(ctx)
	if err != nil {
		t.Error(err) // but try to clean up the prepared txn
	}
	conn, err = pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	closed, err = rollbackPreparedTxns(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	if closed != 1 {
		t.Errorf("rolled back %d, wanted %d", closed, 1)
	}
}

// TestSelectLiteralType ensures (and demonstrates) that simpler query execution
// modes can effectively handle inline queries like `SELECT $1;` when provided
// and argument that is not a string, which fails in the expanded execution
// modes that try to obtain argument OIDs (postgres data types) via a
// prepare/describe request to the postgres process. However, the to get the
// returned type correct and to deal with the invalidity of a statement like
// `SELECT $1 + $2` with text arguments, we provide a special
// QueryModeInferredArgTypes mode so such statements can succeed.
func TestSelectLiteralType(t *testing.T) {
	ctx := context.Background()
	connStr := connString(cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DBName, false)
	pgCfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		t.Fatal(err)
	}
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement // default
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeExec
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(ctx, pgCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(ctx); err != nil {
			t.Error(err)
		}
	})

	err = registerTypes(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}

	// var arg any = int64(1)
	// args := []any{arg, arg}
	argMap := map[string]any{
		"a": int64(1),
		"b": int64(1),
	}
	var arg any = pgx.NamedArgs(argMap)
	args := []any{arg}
	stmt := `SELECT @a + @b;`

	// args := []any{arg}
	// stmt := `SELECT $1;`
	// TODO: make more thorough in-line expression test cases

	rows, err := queryImpliedArgTypes(ctx, conn, stmt, args...)
	if err != nil {
		t.Fatal(err)
	}

	// rows, err := conn.Query(ctx, stmt, arg, arg)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// Tip: try the named argument rewriter as such:
	// argMap := map[string]any{
	// 	"a": int64(1),
	// }
	// var arg any = pgx.NamedArgs(argMap)
	//  ^ with `SELECT @a;`
	// (with the same result)

	defer rows.Close()

	for rows.Next() {
		// rows.Values() // []any
		var val any
		err = rows.Scan(&val)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%v (%T)", val, val)
		// valInt, err := conv.Int(val)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// t.Log(valInt, "(int64)")
	}

	err = rows.Err()
	if err != nil {
		t.Fatal(err)
	}

	// Now with our high level func and mode.
	args2 := append([]any{QueryModeInferredArgTypes}, args...)
	results, err := query(ctx, oidTypesMap(conn.TypeMap()), &cqWrapper{conn}, stmt, args2...)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range results.Rows[0] {
		t.Logf("%v (%T)", val, val)
	}
}

func TestNestedTx(t *testing.T) {
	ctx := context.Background()

	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Query(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	// Start the outer transaction.
	tx, err := db.BeginOuterTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	// err = tx.Commit(ctx)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// tx.begintx below would then error with "tx is closed" (TODO: test that)

	// Start savepoint 0
	txNested, err := tx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer txNested.Rollback(ctx)

	// OK query
	_, err = txNested.Execute(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	// release savepoint 0
	err = txNested.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Start savepoint 1
	txNested2, err := tx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Query error
	_, err = txNested2.Execute(ctx, `SELECT notathing;`)
	if err == nil {
		t.Fatal("should have errored") // expect error
	}
	// if Commit now, should say: ERROR: current transaction is aborted, commands ignored until end of transaction block (SQLSTATE 25P02)

	// rollback savepoint 1 containing the error
	err = txNested2.Rollback(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Now we should be able to keep going.  Make savepoint 3:
	txNested3, err := tx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer txNested3.Rollback(ctx)

	_, err = txNested3.Execute(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	err = txNested3.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	id, err := tx.Precommit(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("commit id: %x", id)

	// success on outer tx even though failure in a savepoint
	err = tx.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: enure updates in other non-failed savepoints take
}

// func TestCommitWithoutPrecommit

// tests that a read tx can be created and used
// while another tx is in progress
func TestReadTxs(t *testing.T) {
	ctx := context.Background()

	db, err := NewDB(ctx, cfg)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Query(ctx, pingStmt)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)

	tx2, err := db.BeginReadTx(ctx)
	require.NoError(t, err)
	defer tx2.Rollback(ctx)

	err = tx.Rollback(ctx)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)
}

// TestTypeRoundtrip tests roundtripping different data types to and from Postgres.
func TestTypeRoundtrip(t *testing.T) {
	type testcase struct {
		// typ specifies the postgres type to use in the test.
		typ string
		// val is the value to pass to the query.
		val any
		// want is the expected value to be returned from the query.
		// if nil, val is expected to be returned
		want any
		// skipInferred skips the inferred arg types test.
		skipInferred bool
		// skipTbl skips the table test. This is used if we are testing
		// a value that isn't directly applicable to postgres.
		skipTbl bool
	}

	for _, v := range []testcase{
		{
			typ: "int8",
			val: int64(1),
		},
		{
			typ: "bool",
			val: true,
		},
		{
			typ: "text",
			val: "hello",
		},
		{
			typ: "bytea",
			val: []byte("world"),
		},
		{
			typ: "uuid",
			val: types.NewUUIDV5([]byte("1")),
		},
		{
			typ: "decimal(6,3)",
			val: mustDecimal("123.456"),
		},
		{
			typ: "decimal(5,0)",
			val: mustDecimal("12300"),
		},
		{
			typ:  "decimal(3,0)",
			val:  types.Uint256FromInt(100),
			want: mustDecimal("100"),
		},
		{
			typ:  "uint256",
			val:  types.Uint256FromInt(100),
			want: mustDecimal("100"),
		},
		{
			typ:  "int8[]",
			val:  []int64{1, 2, 3},
			want: []int64{int64(1), int64(2), int64(3)},
		},
		{
			typ:  "bool[]",
			val:  []bool{true, false, true},
			want: []bool{true, false, true},
		},
		{
			typ:  "text[]",
			val:  []string{"a", "b", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			typ:  "bytea[]",
			val:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			want: [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		},
		{
			typ:  "uuid[]",
			val:  types.UUIDArray{types.NewUUIDV5([]byte("2")), types.NewUUIDV5([]byte("3"))},
			want: types.UUIDArray{types.NewUUIDV5([]byte("2")), types.NewUUIDV5([]byte("3"))},
		},
		{
			typ:  "decimal(6,4)[]",
			val:  decimal.DecimalArray{mustDecimal("12.4223"), mustDecimal("22.4425"), mustDecimal("23.7423")},
			want: decimal.DecimalArray{mustDecimal("12.4223"), mustDecimal("22.4425"), mustDecimal("23.7423")},
		},
		{
			typ:  "uint256[]",
			val:  types.Uint256Array{types.Uint256FromInt(100), types.Uint256FromInt(200), types.Uint256FromInt(300)},
			want: types.Uint256Array{types.Uint256FromInt(100), types.Uint256FromInt(200), types.Uint256FromInt(300)},
		},
		{
			typ:  "text[]",
			val:  []string{},
			want: []string{},
		},
		{
			typ:  "int8[]",
			val:  []int64{},
			want: []int64{},
		},
		{
			typ:     "nil",
			val:     nil,
			skipTbl: true,
		},
		{
			typ:     "[]uuid",
			val:     []any{"3146857c-8671-4f4e-99bd-fcc621f9d3d1", "3146857c-8671-4f4e-99bd-fcc621f9d3d1"},
			want:    []string{"3146857c-8671-4f4e-99bd-fcc621f9d3d1", "3146857c-8671-4f4e-99bd-fcc621f9d3d1"},
			skipTbl: true,
		},
		{
			typ:          "int8[]",
			val:          []string{"1", "2"},
			want:         []int64{int64(1), int64(2)},
			skipInferred: true,
		},
	} {
		t.Run(v.typ, func(t *testing.T) {
			ctx := context.Background()
			db, err := NewDB(ctx, cfg)
			require.NoError(t, err)
			defer db.Close()

			want := v.val
			if v.want != nil {
				want = v.want
			}

			if !v.skipInferred {
				res, err := db.Query(ctx, "SELECT $1", QueryModeInferredArgTypes, v.val)
				require.NoError(t, err)

				require.Len(t, res.Columns, 1)
				require.Len(t, res.Rows, 1)
				require.Len(t, res.Rows[0], 1)

				require.EqualValues(t, want, res.Rows[0][0])
			}

			if v.skipTbl {
				return
			}

			// here, we test without the QueryModeInferredArgTypes

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback

			_, err = tx.Execute(ctx, "CREATE TEMP TABLE test (val "+v.typ+")", QueryModeExec)
			require.NoError(t, err)

			_, err = tx.Execute(ctx, "INSERT INTO test (val) VALUES ($1)", QueryModeExec, v.val)
			require.NoError(t, err)

			res, err := tx.Execute(ctx, "SELECT val FROM test", QueryModeExec)
			require.NoError(t, err)

			require.Len(t, res.Columns, 1)
			require.Len(t, res.Rows, 1)
			require.Len(t, res.Rows[0], 1)

			require.EqualValues(t, want, res.Rows[0][0])
		})
	}
}

// mustDecimal panics if the string cannot be converted to a decimal.
func mustDecimal(s string) *decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustParseUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// mustUint256 panics if the string cannot be converted to a Uint256.
func mustUint256(s string) *types.Uint256 {
	u, err := types.Uint256FromString(s)
	if err != nil {
		panic(err)
	}
	return u
}

func Test_DelayedTx(t *testing.T) {
	ctx := context.Background()

	db, err := NewDB(ctx, cfg)
	require.NoError(t, err)
	defer db.Close()

	tx := db.BeginDelayedReadTx()
	defer tx.Rollback(ctx)

	tx2, err := tx.BeginTx(ctx)
	require.NoError(t, err)
	defer tx2.Rollback(ctx)

	_, err = tx2.Execute(ctx, pingStmt)
	require.NoError(t, err)

	err = tx2.Commit(ctx)
	require.NoError(t, err)
}

// This test tests changesets, and that they are properly encoded+decoded
func Test_Changesets(t *testing.T) {
	for i, tc := range []interface {
		run(t *testing.T)
	}{
		&changesetTestcase[string, []string]{ // basic string test
			datatype:  "text",
			val:       "hello",
			arrayVal:  []string{"a", "b", "c"},
			val2:      "world",
			arrayVal2: []string{"d", "e", "f"},
		},
		&changesetTestcase[string, []string]{ // test with special characters and escaping
			datatype:  "text",
			val:       "heldcsklk;le''\"';",
			arrayVal:  []string{"hel,dcsklk;le','\",';", `";\\sdsw,"''"\',\""`},
			val2:      "world",
			arrayVal2: []string{"'\"", "heldcsklk;le''\"';"},
		},
		&changesetTestcase[int64, []int64]{
			datatype:  "int8",
			val:       1,
			arrayVal:  []int64{1, 2, 3987654},
			val2:      2,
			arrayVal2: []int64{3, 4, 5},
		},
		&changesetTestcase[bool, []bool]{
			datatype:  "bool",
			val:       true,
			arrayVal:  []bool{true, false, true},
			val2:      false,
			arrayVal2: []bool{false, true, false},
		},
		&changesetTestcase[[]byte, [][]byte]{
			datatype:  "bytea",
			val:       []byte("hello"),
			arrayVal:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			val2:      []byte("world"),
			arrayVal2: [][]byte{[]byte("d"), []byte("e"), []byte("f")},
		},
		&changesetTestcase[*decimal.Decimal, decimal.DecimalArray]{
			datatype:  "decimal(6,3)",
			val:       mustDecimal("123.456"),
			arrayVal:  decimal.DecimalArray{mustDecimal("123.456"), mustDecimal("123.456"), mustDecimal("123.456")},
			val2:      mustDecimal("123.457"),
			arrayVal2: decimal.DecimalArray{mustDecimal("123.457"), mustDecimal("123.457"), mustDecimal("123.457")},
		},
		&changesetTestcase[*types.UUID, types.UUIDArray]{
			datatype:  "uuid",
			val:       mustParseUUID("3146857c-8671-4f4e-99bd-fcc621f9d3d1"),
			arrayVal:  types.UUIDArray{mustParseUUID("3146857c-8671-4f4e-99bd-fcc621f9d3d1"), mustParseUUID("3146857c-8671-4f4e-99bd-fcc621f9d3d1")},
			val2:      mustParseUUID("3146857c-8671-4f4e-99bd-fcc621f9d3d2"),
			arrayVal2: types.UUIDArray{mustParseUUID("3146857c-8671-4f4e-99bd-fcc621f9d3d2"), mustParseUUID("3146857c-8671-4f4e-99bd-fcc621f9d3d2")},
		},
		&changesetTestcase[*types.Uint256, types.Uint256Array]{
			datatype:  "uint256",
			val:       mustUint256("18446744073709551615000000"),
			arrayVal:  types.Uint256Array{mustUint256("184467440737095516150000002"), mustUint256("184467440737095516150000001")},
			val2:      mustUint256("18446744073709551615000001"),
			arrayVal2: types.Uint256Array{mustUint256("184467440737095516150000012"), mustUint256("1844674407370955161500000123")},
		},
	} {
		t.Run(fmt.Sprint(i), tc.run)
	}
}

// this is a hack to use generics in the test
type changesetTestcase[T any, T2 any] struct {
	datatype string // the postgres datatype to test
	// the first vals will be inserted.
	// val will be the primary key
	val      T  // the value to test
	arrayVal T2 // the array value to test
	// the second vals will update the first vals
	val2      T  // the second value to test
	arrayVal2 T2 // the second array value to test
}

func (c *changesetTestcase[T, T2]) run(t *testing.T) {
	ctx := context.Background()

	db, err := NewDB(ctx, cfg)
	require.NoError(t, err)
	defer db.Close()

	cleanup := func() {
		db.AutoCommit(true)
		_, err = db.Execute(ctx, "drop table if exists ds_test.test", QueryModeExec)
		require.NoError(t, err)
		_, err = db.Execute(ctx, "drop schema if exists ds_test", QueryModeExec)
		db.AutoCommit(false)
	}
	// attempt to clean up any old failed tests
	cleanup()
	defer cleanup()

	regularTx, err := db.BeginOuterTx(ctx)
	require.NoError(t, err)
	defer regularTx.Rollback(ctx)

	_, err = regularTx.Execute(ctx, "create schema ds_test", QueryModeExec)
	require.NoError(t, err)

	_, err = regularTx.Execute(ctx, "create table ds_test.test (val "+c.datatype+" primary key, array_val "+c.datatype+"[])", QueryModeExec)
	require.NoError(t, err)

	err = regularTx.Commit(ctx)
	require.NoError(t, err)

	/*
		Block 1: Insert
	*/

	writer := new(bytes.Buffer)

	tx, err := db.BeginOuterTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, "insert into ds_test.test (val, array_val) values ($1, $2)", QueryModeExec, c.val, c.arrayVal)
	require.NoError(t, err)

	// get the changeset
	_, err = tx.Precommit(ctx, writer)
	require.NoError(t, err)

	cs, err := DeserializeChangeset(writer.Bytes())
	require.NoError(t, err)

	require.Len(t, cs.Changesets, 1)
	require.Len(t, cs.Changesets[0].Inserts, 1)

	insertVals, err := cs.Changesets[0].DecodeTuple(cs.Changesets[0].Inserts[0])
	require.NoError(t, err)

	// verify the insert vals are equal to the first vals
	require.EqualValues(t, c.val, insertVals[0])
	require.EqualValues(t, c.arrayVal, insertVals[1])

	err = tx.Commit(ctx)
	require.NoError(t, err)

	/*
		Block 2: Update
	*/

	writer = new(bytes.Buffer)

	tx, err = db.BeginOuterTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, "update ds_test.test set val = $1, array_val = $2", QueryModeExec, c.val2, c.arrayVal2)
	require.NoError(t, err)

	_, err = tx.Precommit(ctx, writer)
	require.NoError(t, err)

	cs, err = DeserializeChangeset(writer.Bytes())
	require.NoError(t, err)

	require.Len(t, cs.Changesets, 1)
	require.Len(t, cs.Changesets[0].Updates, 1)

	oldVals, err := cs.Changesets[0].DecodeTuple(cs.Changesets[0].Updates[0][0])
	require.NoError(t, err)

	newVals, err := cs.Changesets[0].DecodeTuple(cs.Changesets[0].Updates[0][1])
	require.NoError(t, err)

	// verify the old vals are equal to the first vals
	require.EqualValues(t, c.val, oldVals[0])
	require.EqualValues(t, c.arrayVal, oldVals[1])

	// verify the new vals are equal to the second vals
	require.EqualValues(t, c.val2, newVals[0])
	require.EqualValues(t, c.arrayVal2, newVals[1])

	err = tx.Commit(ctx)
	require.NoError(t, err)

	/*
		Block 3: Delete
	*/

	writer = new(bytes.Buffer)

	tx, err = db.BeginOuterTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, "delete from ds_test.test", QueryModeExec)
	require.NoError(t, err)

	_, err = tx.Precommit(ctx, writer)
	require.NoError(t, err)

	cs, err = DeserializeChangeset(writer.Bytes())
	require.NoError(t, err)

	require.Len(t, cs.Changesets, 1)
	require.Len(t, cs.Changesets[0].Deletes, 1)

	deleteVals, err := cs.Changesets[0].DecodeTuple(cs.Changesets[0].Deletes[0])
	require.NoError(t, err)

	// verify the delete vals are equal to the second vals
	require.EqualValues(t, c.val2, deleteVals[0])
	require.EqualValues(t, c.arrayVal2, deleteVals[1])

	err = tx.Commit(ctx)
	require.NoError(t, err)
}
