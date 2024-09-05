//go:build pglive

package integration_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test is used to easily test procedure inputs/outputs and logic.
// All tests are given the same schema with a few tables and procedures, as well
// as mock data. The test is then able to define its own procedure, the inputs,
// outputs, and expected error (if any).
func Test_Procedures(t *testing.T) {
	type testcase struct {
		name      string
		procedure string
		inputs    []any    // can be nil
		outputs   [][]any  // can be nil
		err       error    // can be nil
		caller    string   // can be empty, if set it will override the default caller in the transaction data
		readOnly  bool     // if true, the procedure will be executed in a read-only transaction
		notices   []string // expected notices, if any
	}

	tests := []testcase{
		{
			name: "basic test",
			procedure: `procedure create_user2($name text, $usernum int) public {
				INSERT INTO users (id, name, wallet_address, user_num)
				VALUES (uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
					$name,
					@caller,
					$usernum
				);
			}`,
			inputs: []any{"test_user", 4},
		},
		{
			name: "for loop",
			procedure: `procedure get_all_users($ints int[]) public view returns (ints int[]) {
				$result int[];
				for $i in $ints {
					$result := array_append($result, $i*2);
				}
				return $result;
			}
				`,
			inputs:  []any{[]int64{1, 2, 3}},
			outputs: [][]any{{[]int64{int64(2), int64(4), int64(6)}}}, // returns 1 row, 1 column, with an array of ints
		},
		{
			name: "is (null)",
			procedure: `procedure is_null($val text) public view returns (is_null bool, is_null2 bool, is_null3 bool, is_null4 bool) {
				$val2 := 1;
				return $val is not distinct from null, $val2 is not distinct from null, $val is distinct from null, $val2 is distinct from null;
			}`,
			inputs:  []any{nil},
			outputs: [][]any{{true, false, false, true}},
		},
		{
			name: "is (concrete)",
			procedure: `procedure is_equal() public view returns (is_equal bool, is_equal2 bool, is_equal3 bool, is_equal4 bool) {
				$val := 'hello';
				return $val is not distinct from 'hello', $val is not distinct from 'world', $val is distinct from 'hello', $val is distinct from 'world';
			}`,
			outputs: [][]any{{true, false, false, true}},
		},
		{
			name: "equals",
			procedure: `procedure equals($val text) public view returns (is_equal bool, is_equal2 bool, is_equal3 bool, is_equal4 bool) {
				$val2 text;
				return $val = 'hello', $val = 'world', $val != null, $val2 != null;
			}`,
			inputs:  []any{"hello"},
			outputs: [][]any{{true, false, nil, nil}}, // equals with null should return null
		},
		{
			name: "and/or",
			procedure: `procedure and_or() public view returns (count int) {
				$count := 0;
				if true and true {
					$count := $count + 1;
				}
				if true and false {
					$count := $count + 100;
				}

				if (true or false) or (true or true) {
					$count := $count + 10;
				}

				return $count;
			}`,
			outputs: [][]any{{int64(11)}},
		},
		{
			name: "return next from a non-table",
			procedure: `procedure return_next($vals int[]) public view returns table(val int) {
				for $i in $vals {
					return next $i*2;
				}
			}`,
			inputs:  []any{[]int64{1, 2, 3}},
			outputs: [][]any{{int64(2)}, {int64(4)}, {int64(6)}},
		},
		{
			name: "table return with no hits doesn't return postgres no-return error",
			procedure: `procedure return_next($vals int[]) public view returns table(val int) {
				for $i in $vals {
					error('unreachable');
				}
			}`,
			inputs:  []any{[]int64{}},
			outputs: [][]any{},
		},
		{
			name: "loop over null array",
			procedure: `procedure loop_over_null() public view returns (count int) {
				$vals int[];
				$count := 0;
				for $i in $vals {
					$count := $count + 1;
				}
				return $count;
			}`,
			outputs: [][]any{{int64(0)}},
		},
		{
			name: "encode, decode, and digest functions",
			procedure: `procedure encode_decode_digest($hex text) public view returns (encoded text, decoded blob, digest blob) {
				$decoded := decode($hex, 'hex');
				$encoded := encode($decoded, 'base64');
				$digest := digest($decoded, 'sha256');
				return $encoded, $decoded, $digest;
			}`,
			inputs:  []any{hex.EncodeToString([]byte("hello"))},
			outputs: [][]any{{base64.StdEncoding.EncodeToString([]byte("hello")), []byte("hello"), crypto.Sha256([]byte("hello"))}},
		},
		{
			name: "join on subquery",
			procedure: `procedure join_on_subquery() public view returns table(name text, content text) {
				return SELECT u.name, p.content FROM users u
				INNER JOIN (select content, user_id from posts) p ON u.id = p.user_id
				WHERE u.name = 'satoshi';
			}`,
			// should come out LIFO, due to default ordering
			outputs: [][]any{
				{"satoshi", "buy $btc to grow laser eyes"},
				{"satoshi", "goodbye world"},
				{"satoshi", "hello world"},
			},
		},
		{
			name: "string functions",
			procedure: `procedure string_funcs() public view {
				$val := 'hello world';
				$val := $val || '!!!';
				$val := upper($val);
				if $val != 'HELLO WORLD!!!' {
					error('upper failed');
				}
				$val := lower($val);
				if $val != 'hello world!!!' {
					error('lower failed');
				}

				if bit_length($val) != 112 {
					error('bit_length failed');
				}
				if char_length($val) != 14 or character_length($val) != 14 or length($val) != 14 {
					error('length failed');
				}
				if octet_length($val) != 14 {
					error('octet_length failed');
				}
				$val := rtrim($val, '!');
				if $val != 'hello world' {
					error('rtrim failed');
				}
				if rtrim($val||' ') != 'hello world' {
					error('rtrim 2 failed');
				}

				$val := ltrim($val, 'h');
				if $val != 'ello world' {
					error('ltrim failed');
				}
				if ltrim(' '||$val) != 'ello world' { // add a space and trim it off
					error('ltrim 2 failed');
				}

				$val := lpad($val, 11, 'h');
				if $val != 'hello world' {
					error('lpad failed');
				}
				if lpad($val, 12) != ' hello world' {
					error('lpad 2 failed');
				}

				$val := rpad($val, 12, '!');
				if $val != 'hello world!' {
					error('rpad failed');
				}
				if rpad($val, 13) != 'hello world! ' {
					error('rpad 2 failed');
				}

				if overlay($val, 'xx', 2, 5) != 'hxxworld!' {
					error('overlay failed');
				}
				if overlay($val, 'xx', 2) != 'hxxlo world!' {
					error('overlay 2 failed');
				}

				if position('world', $val) != 7 {
					error('position failed');
				}
				if substring($val, 7, 5) != 'world' {
					error('substring failed');
				}
				if substring($val, 7) != 'world!' {
					error('substring 2 failed');
				}

				if trim(' ' || $val || ' ') != 'hello world!' {
					error('trim failed');
				}
				if trim('a'||$val||'a', 'a') != 'hello world!' {
					error('trim 2 failed');
				}

				if parse_unix_timestamp('2021-01-01 00:00:00:123456', 'YYYY-MM-DD HH24:MI:SS:US') != 1609459200.123456 {
					error('parse_unix_timestamp failed');
				}

				if format_unix_timestamp(1609459200.123456, 'YYYY-MM-DD HH24:MI:SS:US') != '2021-01-01 00:00:00:123456' {
					error('format_unix_timestamp failed');
				}

				if generate_dbid('aa', decode('B7E2d6DABaf3B0038cFAaf09688Fa104f4409697', 'hex')) != 'xacfa19c2d4af530c6225ea139d611f91e7a55222a362dfd5eb70a826' {
					error('generate_dbid failed');
				}

				// regression test for invalid generated sql
				$dbid := generate_dbid('aa', decode('B7E2d6DABaf3B0038cFAaf09688Fa104f4409697', 'hex'));
				if $dbid != 'xacfa19c2d4af530c6225ea139d611f91e7a55222a362dfd5eb70a826' {
					error('generate_dbid regression test failed');
				}
			}`,
		},
		{
			name: "arrays",
			// all arrays are 1-indexed
			procedure: `procedure array_funcs() public view {
				$arr int[] := [1, 2, 3];
				$arr := array_append($arr, 4);
				if $arr != [1, 2, 3, 4] {
					error('array_append failed');
				}

				$arr2 := $arr[2:4]; // should be [2, 3, 4]
				if $arr2 != [2, 3, 4] {
					error('array slice failed');
				}

				$arr3 := array_prepend(0, $arr);
				if $arr3 != [0, 1, 2, 3, 4] {
					error('array_prepend failed');
				}

				if array_remove($arr3, 3) != [0, 1, 2, 4] {
					error('array_remove failed');
				}

				if array_cat($arr[:2], $arr[4:]) != [1, 2, 4] {
					error('array_cat failed');
				}
			}`,
		},
		{
			name: "min/max",
			procedure: `procedure min_max() public view returns (min int, max int) {
				$max := 0;
				for $row in select max(user_num) as m from users {
					$max := $row.m;
				}
				$min := 0;
				for $row2 in select min(user_num) as m from users {
					$min := $row2.m;
				}
				return $min, $max;
			}`,
			outputs: [][]any{{int64(1), int64(3)}},
		},
		{
			name: "sum",
			// cannot use duplicate function names, so we use sum2
			procedure: `procedure sum2() public view returns (sum decimal(1000,0)) {
				for $row in select sum(user_num) as s from users {
					return $row.s;
				}
			}`,
			outputs: [][]any{{mustDecimal("6", 1000, 0)}},
		},
		{
			name: "decimal array",
			procedure: `procedure decimal_array() public view returns (decimals decimal(2,1)[]) {
				$a := 2.5;
				$b := 3.5;
				$c := $a/$b;
				return [$a, $b, $c];
			}`,
			outputs: [][]any{{decimal.DecimalArray{mustDecimal("2.5", 2, 1), mustDecimal("3.5", 2, 1), mustDecimal("0.7", 2, 1)}}},
		},
		{
			name: "decimal",
			procedure: `procedure d() public view {
					$i := 100.423;
					$j decimal(16,8) := 46728954.23743892;
					$k := $i::decimal(16,8) + $j;
					if $k != 46729054.66043892 {
						error('decimal failed');
					}
					if $k::text != '46729054.66043892' {
						error('decimal text failed');
					}
					if ($k::decimal(16,2))::text != '46729054.66' {
						error('decimal 2 failed');
					}
				}`,
		},
		{
			name: "early empty return",
			procedure: `procedure return_early() public view {
					$exit := true;
					if $exit {
						return;
					}
					error('should not reach here');
				}`,
		},
		{
			name: "private procedure",
			procedure: `procedure private_proc() private view {
					error('should not reach here');
				}`,
			err: execution.ErrPrivate,
		},
		{
			name: "owner procedure - success",
			procedure: `procedure owner_proc() public owner view returns (is_owner bool) {
					return true;
				}`,
			outputs: [][]any{{true}},
		},
		{
			name: "owner procedure - fail",
			procedure: `procedure owner_proc() public owner view returns (is_owner bool) {
					return false;
				}`,
			err:    execution.ErrOwnerOnly,
			caller: "some_other_wallet",
		},
		{
			name: "mutative procedure in read-only tx",
			procedure: `procedure mutative() public {
					return;
				}`,
			err:      execution.ErrMutativeProcedure,
			readOnly: true,
		},
		{
			// this is a regression test for a previous bug
			name: "unary",
			procedure: `procedure unary() public view returns (bool) {
				return !true;
			}`,
			outputs: [][]any{{false}},
		},
		{
			name: "array",
			procedure: `procedure assign_array() public view returns (ints int[]) {
				$arr int[] := [1, 2, 3];

				$arr[2] := 4;
				return $arr;
			}`,
			outputs: [][]any{{[]int64{int64(1), int64(4), int64(3)}}},
		},
		{
			name: "notice",
			procedure: `procedure notice_fn() public {
				for $i in 1..3 {
					notice($i::text);
				}
			}`,
			notices: []string{"1", "2", "3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			global, db, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginPreparedTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			// deploy schema
			dbid := deployAndSeed(t, global, tx, test.procedure)

			// parse out procedure name
			procedureName := parseProcedureName(test.procedure)

			d := txData()
			if test.caller != "" {
				d.Caller = test.caller
				d.Signer = []byte(test.caller)
			}

			var execTx interface {
				sql.Tx
				sql.Subscriber
			} = tx
			if test.readOnly {
				execTx, err = db.BeginReadTx(ctx)
				require.NoError(t, err)
				defer execTx.Rollback(ctx)
			}

			// listen for notices
			notice, done, err := execTx.Subscribe(ctx)
			require.NoError(t, err)
			defer done(ctx)

			var rec []string
			go func() {
				for n := range notice {
					_, notc, err := parse.ParseNotice(n)
					require.NoError(t, err)
					rec = append(rec, notc)
				}
			}()

			// execute test procedure
			res, err := global.Procedure(ctx, execTx, &common.ExecutionData{
				TransactionData: d,
				Dataset:         dbid,
				Procedure:       procedureName,
				Args:            test.inputs,
			})
			if test.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.err)
				return
			}
			require.NoError(t, err)

			require.Len(t, res.Rows, len(test.outputs))

			for i, output := range test.outputs {
				require.Len(t, res.Rows[i], len(output))
				for j, val := range output {
					if dec, ok := val.(*decimal.Decimal); ok {
						received := res.Rows[i][j].(*decimal.Decimal)

						assert.Equal(t, dec.String(), received.String())
						assert.Equal(t, dec.Precision(), received.Precision())
						assert.Equal(t, dec.Scale(), received.Scale())
						continue
					}

					require.Equal(t, val, res.Rows[i][j])
				}
			}

			// check notices
			require.Equal(t, test.notices, rec)
		})
	}
}

func mustDecimal(val string, precision, scale uint16) *decimal.Decimal {
	d, err := decimal.NewExplicit(val, precision, scale)
	if err != nil {
		panic(err)
	}

	return d
}

func Test_ForeignProcedures(t *testing.T) {
	type testcase struct {
		name string
		// foreign is the foreign procedure definition.
		// It will be deployed in a separate schema.
		foreign string
		// otherProc is the procedure that calls the foreign procedure.
		// It will be included with the foreign procedure.
		// It should be formattable to allow the caller to format with
		// the target dbid, and the target procedure should be hardcoded.
		otherProc string
		// inputs are the inputs to the test procedure.
		inputs []any
		// outputs are the expected outputs from the test procedure.
		outputs [][]any
		// caller is the calling address. If empty, it defaults to the package default.
		caller string
		// if wantErr is not empty, the test will expect an error containing this string.
		// We use a string, instead go Go's error type, because we are reading errors raised
		// from Postgres, which are strings.
		wantErr string
	}

	tests := []testcase{
		{
			name:    "foreign procedure takes nothing, returns nothing",
			foreign: `foreign procedure do_something()`,
			otherProc: `procedure call_foreign() public {
				do_something['%s', 'delete_users']();
			}`,
		},
		{
			name:    "foreign procedure takes nothing, returns table",
			foreign: `foreign procedure get_users() returns table(id uuid, name text, wallet_address text)`,
			otherProc: `procedure call_foreign() public returns table(username text) {
				return select name as username from get_users['%s', 'get_users']();
			}`,
			outputs: [][]any{
				{"satoshi"},
				{"wendys_drive_through_lady"},
				{"zeus"},
			},
		},
		{
			name:    "foreign procedure takes values, returns values",
			foreign: `foreign procedure id_from_name($name text) returns (id uuid)`,
			otherProc: `procedure call_foreign($name text) public returns (id uuid) {
				return id_from_name['%s', 'id_from_name']($name);
			}`,
			inputs:  []any{"satoshi"},
			outputs: [][]any{{satoshisUUID}},
		},
		{
			name:    "foreign procedure expects no args, implementation expects some",
			foreign: `foreign procedure id_from_name() returns (id uuid)`,
			otherProc: `procedure call_foreign() public returns (id uuid) {
				return id_from_name['%s', 'id_from_name']();
			}`,
			wantErr: `requires 1 arg(s)`,
		},
		{
			name:    "foreign procedure expects args, implementation expects none",
			foreign: `foreign procedure get_users($name text) returns table(id uuid, name text, wallet_address text)`,
			otherProc: `procedure call_foreign() public returns table(username text) {
				return select name as username from get_users['%s', 'get_users']('satoshi');
			}`,
			wantErr: "requires no args",
		},
		{
			name:    "foreign procedure expects 2 args, implementation expects 2",
			foreign: `foreign procedure id_from_name($name text, $name2 text) returns (id uuid)`,
			otherProc: `procedure call_foreign() public returns (id uuid) {
				return id_from_name['%s', 'id_from_name']('satoshi', 'zeus');
			}`,
			wantErr: "requires 1 arg(s)",
		},
		{
			name:    "foreign procedure returns 1 arg, implementation returns none",
			foreign: `foreign procedure delete_users() returns (text)`,
			otherProc: `procedure call_foreign() public returns (text) {
				return delete_users['%s', 'delete_users']();
			}`,
			wantErr: "returns nothing",
		},
		{
			name:    "foreign procedure returns 0 args, implementation returns 1",
			foreign: `foreign procedure id_from_name($name text)`,
			otherProc: `procedure call_foreign() public {
				id_from_name['%s', 'id_from_name']('satoshi');
			}`,
			wantErr: "returns non-nil value(s)",
		},
		{
			name:    "foreign procedure returns table, implementation returns non-table",
			foreign: `foreign procedure id_from_name($name text) returns table(id uuid)`,
			otherProc: `procedure call_foreign() public {
				select id from id_from_name['%s', 'id_from_name']('satoshi');
			}`,
			wantErr: "does not return a table",
		},
		{
			name:    "foreign procedure does not return table, implementation returns table",
			foreign: `foreign procedure get_users() returns (id uuid, name text, wallet_address text)`,
			otherProc: `procedure call_foreign() public returns table(username text) {
				$id, $name, $wallet := get_users['%s', 'get_users']();
			}`,
			wantErr: "returns a table",
		},
		{
			name:    "foreign procedure returns table, implementation returns nothing",
			foreign: `foreign procedure create_user($name text) returns table(id uuid)`,
			otherProc: `procedure call_foreign() public {
				create_user['%s', 'create_user']('satoshi');
			}`,
			wantErr: "does not return a table",
		},
		{
			name: "procedures returning scalar return different named values (ok)",
			// returns value "uid" instead of impl's "id"
			foreign: `foreign procedure id_from_name($name text) returns (uid uuid)`,
			otherProc: `procedure call_foreign() public returns (id uuid) {
				return id_from_name['%s', 'id_from_name']('satoshi');
			}`,
			outputs: [][]any{{satoshisUUID}},
		},
		{
			name:    "procedure returning table return different column names (failure)",
			foreign: `foreign procedure get_users() returns table(uid uuid, name text, wallet_address text)`,
			otherProc: `procedure call_foreign() public returns table(name text) {
				return select name from get_users['%s', 'get_users']();
			}`,
			wantErr: "returns id",
		},
		{
			name:    "private procedure via foreign call",
			foreign: `foreign procedure is_private($name text)`,
			otherProc: `procedure call_foreign() public {
				is_private['%s', 'is_private']('satoshi');
			}`,
			wantErr: "not public",
		},
		// {
		// 	name:    "foreign call owner - fail",
		// 	foreign: `foreign procedure is_owner($name text)`,
		// 	otherProc: `procedure call_foreign() public owner {
		// 		is_owner['%s', 'is_owner']('satoshi');
		// 	}`,
		// 	caller:  "some_other_wallet",
		// 	wantErr: "is owner-only",
		// },
		{
			name:    "foreign call owner - success",
			foreign: `foreign procedure is_owner($name text)`,
			otherProc: `procedure call_foreign() public owner {
				is_owner['%s', 'is_owner']('satoshi');
			}`,
		},
		// this test tests that foreign caller properly works, and is unset at the end of the
		// foreign call.
		{
			name:    "testing foreign caller",
			foreign: `foreign procedure return_foreign_caller() returns (caller text)`,
			otherProc: `procedure call_foreign() public returns (one text, two text, three text) {
			$one := @foreign_caller;
			$two := return_foreign_caller['%s', 'return_foreign_caller']();
			$three := @foreign_caller;

			return $one, $two, $three;
		}`,
			outputs: [][]any{{"", "x93c803781453c866b8e1277d6d13eaa17935d891544fd223e0ea75b0", ""}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			global, db, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			// deploy the main test schema
			foreignDBID := deployAndSeed(t, global, tx)

			// deploy the new schema that will call the main one
			// first, format the procedure with the foreign DBID
			otherProc := fmt.Sprintf(test.otherProc, foreignDBID)

			// deploy the new schema
			mainDBID := deploy(t, global, tx, fmt.Sprintf("database db2;\n%s\n%s", test.foreign, otherProc))

			procedureName := parseProcedureName(otherProc)

			d := txData()
			if test.caller != "" {
				d.Caller = test.caller
				d.Signer = []byte(test.caller)
			}

			// execute test procedure
			res, err := global.Procedure(ctx, tx, &common.ExecutionData{
				TransactionData: d,
				Dataset:         mainDBID,
				Procedure:       procedureName,
				Args:            test.inputs,
			})
			if test.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)

			require.Len(t, res.Rows, len(test.outputs))
			for i, output := range test.outputs {
				require.Len(t, res.Rows[i], len(output))
				for j, val := range output {
					require.Equal(t, val, res.Rows[i][j])
				}
			}
		})
	}
}

// testSchema is a schema that can be deployed with deployAndSeed
var testSchema = `
database ecclesia;

table users {
	id uuid primary key,
	name text not null maxlen(100) minlen(4) unique,
	wallet_address text not null,
	user_num int unique notnull // this could be the primary key, but it's more for testing than to be useful
}

table posts {
	id uuid primary key,
	user_id uuid not null,
	content text not null maxlen(300),
	foreign key (user_id) references users(id) on delete cascade
}

procedure create_user($name text) public {
	$max int;
	for $row in select max(user_num) as m from users {
		$max := $row.m;
	}

	if $max is null  {
		$max := 0;
	}

	INSERT INTO users (id, name, wallet_address, user_num)
	VALUES (uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
		$name,
		@caller,
		$max + 1
	);
}

procedure owns_user($wallet text, $name text) public view returns (owns bool) {
	$exists bool := false;
	for $row in SELECT * FROM users WHERE wallet_address = $wallet
	AND name = $name {
		$exists := true;
	}

	return $exists;
}

procedure id_from_name($name text) public view returns (id uuid) {
	for $row in SELECT id FROM users WHERE name = $name {
		return $row.id;
	}
	error('user not found');
}

procedure create_post($username text, $content text) public {
	if owns_user(@caller, $username) == false {
		error('caller does not own user');
	}

	INSERT INTO posts (id, user_id, content)
	VALUES (uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
		id_from_name($username),
		$content
	);
}

// the following procedures serve no utility, and are made only to test foreign calls
// to different signatures.
procedure delete_users() public {
	DELETE FROM users;
}

procedure get_users() public returns table(id uuid, name text, wallet_address text) {
	return SELECT id, name, wallet_address FROM users;
}

// matches create_user signature
procedure is_private($name text) private {
	error('should not reach here');
}

procedure is_owner($name text) public owner view {
	$exists bool := false;
}

procedure return_foreign_caller() public returns (caller text) {
	return @foreign_caller;
}
`

// maps usernames to post content.
var initialData = map[string][]string{
	"satoshi":                   {"hello world", "goodbye world", "buy $btc to grow laser eyes"},
	"zeus":                      {"i am zeus", "i am the god of thunder", "i am the god of lightning"},
	"wendys_drive_through_lady": {"hi how can I help you", "no I don't know what the federal reserve is", "sir this is a wendys"},
}

var satoshisUUID = &types.UUID{0x38, 0xeb, 0x77, 0xcb, 0x1e, 0x5a, 0x56, 0xc0, 0x85, 0x63, 0x2e, 0x25, 0x34, 0xd6, 0x7b, 0x96}

// deploy deploys a schema.
// if deployer is not "", it will set the deployer as the owner.
func deploy(t *testing.T, global *execution.GlobalContext, db sql.DB, schema string) (dbid string) {
	ctx := context.Background()

	parsed, err := parse.Parse([]byte(schema))
	require.NoError(t, err)

	d := txData()

	err = global.CreateDataset(ctx, db, parsed, &d)
	require.NoError(t, err)

	// get dbid
	dbs, err := global.ListDatasets(owner)
	require.NoError(t, err)

	for _, db := range dbs {
		if db.Name == parsed.Name {
			dbid = db.DBID
			break
		}
	}

	return dbid
}

// deployAndSeed deploys the test schema and seeds it with data
func deployAndSeed(t *testing.T, global *execution.GlobalContext, db sql.DB, extraProcedures ...string) (dbid string) {
	ctx := context.Background()

	schema := testSchema
	for _, proc := range extraProcedures {
		schema += proc + "\n"
	}

	// deploy schema
	dbid = deploy(t, global, db, schema)

	// create initial data
	for _, kv := range order.OrderMap(initialData) {
		_, err := global.Procedure(ctx, db, &common.ExecutionData{
			TransactionData: txData(),
			Dataset:         dbid,
			Procedure:       "create_user",
			Args:            []any{kv.Key},
		})
		require.NoError(t, err)

		for _, post := range kv.Value {
			_, err = global.Procedure(ctx, db, &common.ExecutionData{
				TransactionData: txData(),
				Dataset:         dbid,
				Procedure:       "create_post",
				Args:            []any{kv.Key, post},
			})
			require.NoError(t, err)
		}
	}

	return dbid
}

// parseProcedureName parses the procedure name from a procedure definition
func parseProcedureName(proc string) string {
	procs := strings.Split(proc, " ")
	procedureName := strings.Split(procs[1], "(")[0]
	procedureName = strings.TrimSpace(procedureName)
	return procedureName
}
