//go:build pglive

package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/require"
)

// This test is used to easily test procedure inputs/outputs and logic.
// All tests are given the same schema with a few tables and procedures, as well
// as mock data. The test is then able to define its own procedure, the inputs,
// outputs, and expected error (if any).
func Test_Procedures(t *testing.T) {
	schema := `
	database ecclesia;

	table users {
		id uuid primary key,
		name text not null maxlen(100) minlen(4) unique,
		wallet_address text not null
	}

	table posts {
		id uuid primary key,
		user_id uuid not null,
		content text not null maxlen(300),
		foreign key (user_id) references users(id)
	}

	procedure create_user($name text) public {
		INSERT INTO users (id, name, wallet_address)
		VALUES (uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
			$name,
			@caller
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
	`

	// maps usernames to post content.
	initialData := map[string][]string{
		"satoshi":                   {"hello world", "goodbye world", "buy $btc to grow laser eyes"},
		"zeus":                      {"i am zeus", "i am the god of thunder", "i am the god of lightning"},
		"wendys_drive_through_lady": {"hi how can I help you", "no I don't know what the federal reserve is", "sir this is a wendys"},
	}

	type testcase struct {
		name      string
		procedure string
		inputs    []any   // can be nil
		outputs   [][]any // can be nil
		err       error   // can be nil
	}

	tests := []testcase{
		{
			name: "basic test",
			procedure: `procedure create_user2($name text) public {
				INSERT INTO users (id, name, wallet_address)
				VALUES (uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
					$name,
					@caller
				);
			}`,
			inputs: []any{"test_user"},
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
			outputs: [][]any{{[]any{int64(2), int64(4), int64(6)}}}, // returns 1 row, 1 column, with an array of ints
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

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			// deploy schema
			parsed, err := parse.ParseSchema([]byte(schema + test.procedure))
			require.NoError(t, err)
			require.NoError(t, parsed.Err())

			err = global.CreateDataset(ctx, tx, parsed.Schema, &common.TransactionData{
				Signer: []byte("deployer"),
				Caller: "deployer",
				TxID:   "deploydb",
			})
			require.NoError(t, err)

			// get dbid
			dbs, err := global.ListDatasets([]byte("deployer"))
			require.NoError(t, err)
			require.Len(t, dbs, 1)
			dbid := dbs[0].DBID

			// create initial data
			for username, posts := range initialData {
				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: []byte("username_signer"),
						Caller: "username_caller",
						TxID:   "create_user_" + username,
					},
					Dataset:   dbid,
					Procedure: "create_user",
					Args:      []any{username},
				})
				require.NoError(t, err)

				for i, post := range posts {
					_, err = global.Procedure(ctx, tx, &common.ExecutionData{
						TransactionData: common.TransactionData{
							Signer: []byte("username_signer"),
							Caller: "username_caller",
							TxID:   "create_post_" + username + "_" + fmt.Sprint(i),
						},
						Dataset:   dbid,
						Procedure: "create_post",
						Args:      []any{username, post},
					})
					require.NoError(t, err)
				}
			}

			// parse out procedure name
			procs := strings.Split(test.procedure, " ")
			procedureName := strings.Split(procs[1], "(")[0]
			procedureName = strings.TrimSpace(procedureName)

			// execute test procedure
			res, err := global.Procedure(ctx, tx, &common.ExecutionData{
				TransactionData: common.TransactionData{
					Signer: []byte("test_signer"),
					Caller: "test_caller",
					TxID:   "test",
				},
				Dataset:   dbid,
				Procedure: procedureName,
				Args:      test.inputs,
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
					require.Equal(t, val, res.Rows[i][j])
				}
			}
		})
	}
}
