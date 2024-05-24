//go:build pglive

package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/require"
)

var (
	owner = []byte("test_owner")
)

// Test_Schemas is made to test full kuneiform schemas against the engine.
// The intent of this is to test the full engine, with expected error messages,
// without having to write a full integration test.
func Test_Schemas(t *testing.T) {
	type testCase struct {
		name string
		// fn is the test function
		// the passed db will be in a transaction
		fn func(t *testing.T, global *execution.GlobalContext, db sql.DB)
	}

	// the tests rely on three schemas:
	// users: a table of users, which maps a wallet address to a human readable name
	// social_media: a table of posts and post_counts. posts contains posts, and post_counts contains the number of posts a user has made.
	// video_game: a table of scores tracks users high scores in a video game.
	// posts and video+game also have admin commands for setting the dbid and procedure names.
	testCases := []testCase{
		{
			name: "create user, make several posts, and get posts",
			fn: func(t *testing.T, global *execution.GlobalContext, db sql.DB) {
				usersDBID, socialDBID, _ := deployAllSchemas(t, global, db)
				_ = socialDBID

				ctx := context.Background()

				// create user
				_, err := global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         usersDBID,
					Procedure:       "create_user",
					Args:            []any{"satoshi"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// make a post
				_, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         socialDBID,
					Procedure:       "create_post",
					Args:            []any{"hello world"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// make another post
				_, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         socialDBID,
					Procedure:       "create_post",
					Args:            []any{"goodbye world"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// make one more large post
				_, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         socialDBID,
					Procedure:       "create_post",
					Args:            []any{"this is a longer post than the others`"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// get posts using get_recent_posts
				res, err := global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         socialDBID,
					Procedure:       "get_recent_posts",
					Args:            []any{"satoshi"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// check the columns. should be id, content
				require.Len(t, res.Columns, 2)
				require.Equal(t, "id", res.Columns[0])
				require.Equal(t, "content", res.Columns[1])

				// check the values
				// the last post should be the first one returned
				require.Len(t, res.Rows, 3)
				require.Equal(t, "this is a longer post than the others`", res.Rows[0][1])
				require.Equal(t, "goodbye world", res.Rows[1][1])
				require.Equal(t, "hello world", res.Rows[2][1])

				// use get_recent_posts_by_size to only get posts larger than 20 characters
				res, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         socialDBID,
					Procedure:       "get_recent_posts_by_size",
					Args:            []any{"satoshi", 20, 10}, // takes username, size, limit
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// check the columns. should be id, content
				require.Len(t, res.Columns, 2)
				require.Equal(t, "id", res.Columns[0])
				require.Equal(t, "content", res.Columns[1])

				// check the values
				// the last post should be the first one returned
				require.Len(t, res.Rows, 1)
				require.Equal(t, "this is a longer post than the others`", res.Rows[0][1])
			},
		},
		{
			// video game schema contains other functionalities, such as type assertions
			// arithmetic, etc. TODO: add here once we support fixed point arithmetic
			name: "test video game schema",
			fn: func(t *testing.T, global *execution.GlobalContext, db sql.DB) {
				usersDBID, _, gameDBID := deployAllSchemas(t, global, db)

				ctx := context.Background()

				// create user
				_, err := global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         usersDBID,
					Procedure:       "create_user",
					Args:            []any{"satoshi"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// set the user's high score
				_, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         gameDBID,
					Procedure:       "set_high_score",
					Args:            []any{100},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// get the user's high score
				res, err := global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:         gameDBID,
					Procedure:       "get_high_score",
					Args:            []any{"satoshi"},
					TransactionData: txData(),
				})
				require.NoError(t, err)

				// check the columns. should be score, as an int
				require.Len(t, res.Columns, 1)
				require.Equal(t, "score", res.Columns[0])

				// check the values
				require.Len(t, res.Rows, 1)
				require.Equal(t, int64(100), res.Rows[0][0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			global, db, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			tc.fn(t, global, tx)
		})
	}
}

// loadSchema loads a schema from the schemas directory.
func loadSchema(file string) (*types.Schema, error) {
	d, err := os.ReadFile("./schemas/" + file)
	if err != nil {
		return nil, err
	}

	db, err := parse.ParseAndValidate(d)
	if err != nil {
		return nil, err
	}

	if db.Err() != nil {
		return nil, db.Err()
	}

	return db.Schema, nil
}

// deployAllSchemas deploys all schemas in the schemas directory.
// it returns the dbid of the deployed schemas.
// It will also properly configure the metadata for social_media and video_game.
func deployAllSchemas(t *testing.T, global *execution.GlobalContext, db sql.DB) (usersDBID, socialMediaDBID, videoGameDBID string) {
	ctx := context.Background()
	schemas := []string{"users.kf", "social_media.kf", "video_game.kf"}
	for _, schema := range schemas {
		schema, err := loadSchema(schema)
		require.NoError(t, err)

		transactionData := txData()
		err = global.CreateDataset(ctx, db, schema, &transactionData)
		require.NoError(t, err)
	}

	datasets, err := global.ListDatasets(owner)
	require.NoError(t, err)

	// get the dbids for the three datasets
	var users, socialMedia, videoGame string
	for _, dataset := range datasets {
		switch dataset.Name {
		case "users":
			users = dataset.DBID
		case "social_media":
			socialMedia = dataset.DBID
		case "video_game":
			videoGame = dataset.DBID
		}
	}
	require.NotEmpty(t, users)
	require.NotEmpty(t, socialMedia)
	require.NotEmpty(t, videoGame)

	// set the metadata for social_media and video_game
	// they each need three types of metadata:
	// - dbid: the dbid of the dataset
	// - userbyname: the procedure to get a user by name
	// - userbyowner: the procedure to get a user by owner
	for _, dbid := range []string{socialMedia, videoGame} {
		_, err := global.Procedure(ctx, db, &common.ExecutionData{
			Dataset:         dbid,
			Procedure:       "admin_set",
			Args:            []any{"dbid", users},
			TransactionData: txData(),
		})
		require.NoError(t, err)

		_, err = global.Procedure(ctx, db, &common.ExecutionData{
			Dataset:         dbid,
			Procedure:       "admin_set",
			Args:            []any{"userbyname", "get_user_by_name"},
			TransactionData: txData(),
		})
		require.NoError(t, err)

		_, err = global.Procedure(ctx, db, &common.ExecutionData{
			Dataset:         dbid,
			Procedure:       "admin_set",
			Args:            []any{"userbyowner", "get_user_by_owner"},
			TransactionData: txData(),
		})
		require.NoError(t, err)
	}

	return users, socialMedia, videoGame
}

// txCounter is a global counter for transaction ids.
var txCounter int

func nextTxID() string {
	txCounter++
	return fmt.Sprintf("tx_%d", txCounter)
}

// txData returns a common.TransactionData with the owner as the signer and caller.
func txData() common.TransactionData {
	return common.TransactionData{
		Signer: owner,
		Caller: string(owner),
		TxID:   nextTxID(),
	}
}
