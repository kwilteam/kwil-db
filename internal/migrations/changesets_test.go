package migrations

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangesetMigration(t *testing.T) {
	ctx := context.Background()

	db, err := dbtest.NewTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	cleanup := func() {
		fmt.Println("Cleaning up")
		db.AutoCommit(true)
		_, err = db.Execute(ctx, "drop table if exists ds_test.test", pg.QueryModeExec)
		require.NoError(t, err)
		_, err = db.Execute(ctx, "drop schema if exists ds_test", pg.QueryModeExec)
		require.NoError(t, err)
		_, err = db.Execute(ctx, `DROP SCHEMA IF EXISTS `+migrationsSchemaName+` CASCADE;`)
		require.NoError(t, err)
		db.AutoCommit(false)
	}
	// attempt to clean up any old failed tests
	cleanup()
	defer cleanup()

	err = createTestSchema(ctx, db, t)
	require.NoError(t, err)

	bts, err := sampleChangeset(ctx, db, t)
	require.NoError(t, err)

	fmt.Println(len(bts))
	// Split the changeset into chunks of 100 bytes each

	logger := log.NewStdOut(log.InfoLevel)

	migrator, err = SetupMigrator(ctx, db, nil, nil, "migration_test", logger)
	require.NoError(t, err)

	// Create a changeset migration
	height := big.NewInt(1)
	var csMigrations []*ChangesetMigration
	idx := 1
	totalChunks := len(bts) / 100
	if len(bts)%100 != 0 {
		totalChunks++
	}

	for i := 0; i < len(bts); i += 100 {
		end := i + 100
		if end > len(bts) {
			end = len(bts)
		}
		csMigrations = append(csMigrations, &ChangesetMigration{
			Height:      height,
			ChunkIdx:    big.NewInt(int64(idx)),
			TotalChunks: big.NewInt(int64(totalChunks)),
			Changeset:   bts[i:end],
		})
		idx++
	}

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Insert the changeset migration into the database
	for _, cs := range csMigrations {
		nestedTx, err := tx.BeginTx(ctx)
		require.NoError(t, err)

		// check if all changesets are received for this height
		allReceived, err := allChunksReceived(ctx, nestedTx, cs.Height.Int64())
		require.NoError(t, err)
		assert.False(t, allReceived, "all chunks should not be received")

		// insert the changeset chunk
		err = cs.insertChangeset(ctx, nestedTx)
		require.NoError(t, err)

		require.NoError(t, nestedTx.Commit(ctx))
	}

	// Check if all changesets are received
	allReceived, err := allChunksReceived(ctx, tx, height.Int64())
	require.NoError(t, err)
	require.True(t, allReceived)

	// Get the changesets
	changesets, err := getChangesets(ctx, tx, height.Int64())
	require.NoError(t, err)

	// Check if the changesets are the same
	require.Equal(t, bts, changesets)

	// extract the block changesets info.
	blockChangesets := &BlockChangesets{}
	err = blockChangesets.UnmarshalBinary(bts)
	require.NoError(t, err)

	// apply the changesets
	csGroup := &pg.ChangesetGroup{
		Changesets: blockChangesets.Changesets,
	}
	err = csGroup.ApplyChangesets(ctx, tx)
	require.NoError(t, err)

	tx.Commit(ctx)

	// Check if the changesets were applied
	tx, err = db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	res, err := tx.Execute(ctx, "select * from ds_test.test", pg.QueryModeExec)
	require.NoError(t, err)

	require.Len(t, res.Rows, 2)
	require.Equal(t, int64(1), res.Rows[0][0])
	require.Equal(t, "hello", res.Rows[0][1])
}

func createTestSchema(ctx context.Context, db sql.PreparedTxMaker, t *testing.T) error {
	regularTx, err := db.BeginPreparedTx(ctx)
	require.NoError(t, err)
	defer regularTx.Rollback(ctx)

	_, err = regularTx.Execute(ctx, "create schema ds_test", pg.QueryModeExec)
	require.NoError(t, err)

	_, err = regularTx.Execute(ctx, "create table ds_test.test (val int primary key, name text,  array_val int[])", pg.QueryModeExec)
	require.NoError(t, err)

	err = regularTx.Commit(ctx)
	require.NoError(t, err)
	return err
}

func sampleChangeset(ctx context.Context, db sql.PreparedTxMaker, t *testing.T) ([]byte, error) {
	writer := new(bytes.Buffer)

	tx, err := db.BeginPreparedTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, "insert into ds_test.test (val, name, array_val) values ($1, $2, $3)", pg.QueryModeExec, 1, "hello", []int64{1, 2, 3})
	require.NoError(t, err)

	_, err = tx.Execute(ctx, "insert into ds_test.test (val, name, array_val) values ($1, $2, $3)", pg.QueryModeExec, 2, "mellow", []int64{11, 22, 33})
	require.NoError(t, err)

	_, err = tx.Precommit(ctx, writer)
	require.NoError(t, err)

	cs, err := pg.DeserializeChangeset(writer)
	require.NoError(t, err)

	require.Len(t, cs.Changesets, 1)
	require.Len(t, cs.Changesets[0].Inserts, 2)
	// Rollback the changes
	err = tx.Rollback(ctx)
	require.NoError(t, err)

	blockChangesets := &BlockChangesets{
		Changesets: cs.Changesets,
	}

	// Serialize the changeset
	bts, err := blockChangesets.MarshalBinary()
	require.NoError(t, err)

	return bts, nil
}
