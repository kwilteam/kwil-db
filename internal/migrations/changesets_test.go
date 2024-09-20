//go:build pglive

package migrations

import (
	"bytes"
	"context"
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
	t.Cleanup(func() {
		db.Close()
	})

	cleanup := func() {
		db.AutoCommit(true)
		_, err = db.Execute(ctx, "drop table if exists ds_test.test")
		require.NoError(t, err)
		_, err = db.Execute(ctx, "drop schema if exists ds_test")
		require.NoError(t, err)
		_, err = db.Execute(ctx, `DROP SCHEMA IF EXISTS `+migrationsSchemaName+` CASCADE;`)
		require.NoError(t, err)
		db.AutoCommit(false)
	}
	// attempt to clean up any old failed tests
	cleanup()
	t.Cleanup(func() {
		cleanup()
	})

	err = createTestSchema(ctx, db, t)
	require.NoError(t, err)

	bts, err := sampleChangeset(ctx, db, t)
	require.NoError(t, err)

	// Split the changeset into chunks of 100 bytes each
	logger := log.NewStdOut(log.InfoLevel)

	migrator, err = SetupMigrator(ctx, db, nil, nil, "migration_test", logger)
	require.NoError(t, err)

	// Create a changeset migration
	height := uint64(1)
	var csMigrations []*changesetMigration
	totalChunks := uint64(len(bts) / 100)
	if len(bts)%100 != 0 {
		totalChunks++
	}

	idx := uint64(0)
	for i := 0; i < len(bts); i += 100 {
		end := i + 100
		if end > len(bts) {
			end = len(bts)
		}
		csMigrations = append(csMigrations, &changesetMigration{
			Height:        height,
			ChunkIdx:      idx,
			TotalChunks:   totalChunks,
			Changeset:     bts[i:end],
			PreviousBlock: 0,
		})
		idx++
	}

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var h, ph, total, rcvd int64

	// Insert the changeset migration into the database
	for idx, cs := range csMigrations {
		nestedTx, err := tx.BeginTx(ctx)
		require.NoError(t, err)

		// check if all changesets are received for this height
		h, ph, total, rcvd, err = getEarliestChangesetMetadata(ctx, tx)
		require.NoError(t, err)

		if idx != 0 {
			require.NoError(t, err)
			require.Equal(t, int64(height), h)
			require.Equal(t, int64(0), ph)
			assert.Equal(t, total, int64(totalChunks))
			assert.NotEqual(t, rcvd, idx, "total chunks should not be equal")
		}

		// insert the changeset chunk
		err = cs.insertChangeset(ctx, nestedTx)
		require.NoError(t, err)

		_, err = getChangeset(ctx, nestedTx, int64(height), int64(idx))
		require.NoError(t, err)
		require.NoError(t, nestedTx.Commit(ctx))
	}

	// Check if all changesets are received
	h, ph, total, rcvd, err = getEarliestChangesetMetadata(ctx, tx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), h)
	assert.Equal(t, int64(0), ph)
	assert.Equal(t, total, rcvd, "total chunks should be equal")

	// Apply the changesets
	err = applyChangeset(ctx, tx, int64(height), int64(totalChunks))

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
	tx, err := db.BeginPreparedTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, "insert into ds_test.test (val, name, array_val) values ($1, $2, $3)", pg.QueryModeExec, 1, "hello", []int64{1, 2, 3})
	require.NoError(t, err)

	_, err = tx.Execute(ctx, "insert into ds_test.test (val, name, array_val) values ($1, $2, $3)", pg.QueryModeExec, 2, "mellow", []int64{11, 22, 33})
	require.NoError(t, err)

	changes := make(chan any, 1)
	var changesetEntries []*pg.ChangesetEntry
	var relations []*pg.Relation
	done := make(chan struct{})
	var csbts bytes.Buffer

	go func() {
		defer close(done)
		for ce := range changes {
			switch ce := ce.(type) {
			case *pg.ChangesetEntry:
				changesetEntries = append(changesetEntries, ce)
				err := pg.StreamElement(&csbts, ce)
				if err != nil {
					t.Error(err)
					return
				}

			case *pg.Relation:
				relations = append(relations, ce)
				err := pg.StreamElement(&csbts, ce)
				if err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()

	_, err = tx.Precommit(ctx, changes)
	require.NoError(t, err)

	<-done
	require.Len(t, changesetEntries, 2)
	require.Len(t, relations, 1)

	// Rollback the changes
	err = tx.Rollback(ctx)
	require.NoError(t, err)

	return csbts.Bytes(), nil
}
