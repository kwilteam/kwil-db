package migrations

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common/sql"
)

// InitializeMigrationSchema initializes the migration schema in the database.
func initializeMigrationSchema(ctx context.Context, db sql.DB) error {
	_, err := db.Execute(ctx, tableMigrationsSQL)
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, tableChangesetsMetadataSQL)
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, tableChangesetsSQL)
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, tableLastChangesetSQL)
	if err != nil {
		return err
	}

	return err
}

const (
	migrationsSchemaName   = `kwild_migrations`
	migrationSchemaVersion = 0
)

var (
	// tableMigrationsSQL is the sql table used to store the current migration state.
	// Only one migration can be active at a time.
	// Primary key should always be 1, to help us ensure there are no bugs in the code.
	tableMigrationsSQL = `CREATE TABLE IF NOT EXISTS ` + migrationsSchemaName + `.migration (
		id INT PRIMARY KEY,
		start_height INT NOT NULL,
		end_height INT NOT NULL,
		chain_id TEXT NOT NULL
	)`

	// getMigrationSQL is the sql query used to get the current migration.
	getMigrationSQL = `SELECT start_height, end_height, chain_id FROM ` + migrationsSchemaName + `.migration;`
	// migrationIsActiveSQL is the sql query used  to check if a migration is active.
	migrationIsActiveSQL = `SELECT EXISTS(SELECT 1 FROM ` + migrationsSchemaName + `.migration);`
	// createMigrationSQL is the sql query used to create a new migration.
	createMigrationSQL = `INSERT INTO ` + migrationsSchemaName + `.migration (id, start_height, end_height, chain_id) VALUES ($1, $2, $3, $4);`
)

// getMigrationState gets the current migration state from the database.
func getMigrationState(ctx context.Context, db sql.Executor) (*activeMigration, error) {
	res, err := db.Execute(ctx, getMigrationSQL)
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		return nil, nil
	}

	if len(res.Rows) != 1 {
		// should never happen
		return nil, fmt.Errorf("internal bug: expected one row for migrations, got %d", len(res.Rows))
	}

	// parse the migration declaration
	md := &activeMigration{}

	row := res.Rows[0]
	var ok bool
	md.StartHeight, ok = row[0].(int64)
	if !ok {
		return nil, fmt.Errorf("internal bug: activation period is not an int64")
	}

	md.EndHeight, ok = row[1].(int64)
	if !ok {
		return nil, fmt.Errorf("internal bug: duration is not an int64")
	}

	md.ChainID, ok = row[2].(string)
	if !ok {
		return nil, fmt.Errorf("internal bug: chain ID is not a string")
	}

	return md, nil
}

// migrationActive checks if a migration is active.
func migrationActive(ctx context.Context, db sql.Executor) (bool, error) {
	res, err := db.Execute(ctx, migrationIsActiveSQL)
	if err != nil {
		return false, err
	}

	if len(res.Rows) != 1 {
		// should never happen
		return false, fmt.Errorf("internal bug: expected one row for migrations, got %d", len(res.Rows))
	}

	row := res.Rows[0]
	active, ok := row[0].(bool)
	if !ok {
		return false, fmt.Errorf("internal bug: migration active is not a bool")
	}

	return active, nil
}

// createMigration creates a new migration state in the database.
func createMigration(ctx context.Context, db sql.Executor, md *activeMigration) error {
	_, err := db.Execute(ctx, createMigrationSQL, 1, md.StartHeight, md.EndHeight, md.ChainID)
	return err
}

// Changeset Resolutions
var (
	// Tables concerning migrating changesets from old chain to new chain.
	defaultRowName = `last_changeset`

	// tableLastChangesetSQL is the table that tracks last applied changeset. It is a single row table with the row name as "last_changeset".
	tableLastChangesetSQL = `CREATE TABLE IF NOT EXISTS ` + migrationsSchemaName + `.last_changeset (
		name TEXT PRIMARY KEY,
		height INT
	)`

	// upsertLastChangesetSQL is the sql query used to set the last changeset.
	upsertLastChangesetSQL = `INSERT INTO ` + migrationsSchemaName + `.last_changeset (name, height) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET height = $2;`

	// getLastChangesetSQL is the sql query used to get the last changeset.
	getLastChangesetSQL = `SELECT height FROM ` + migrationsSchemaName + `.last_changeset WHERE name = $1;`

	// tableChangesetMetadataSQL is the table that stores changesets.
	tableChangesetsMetadataSQL = `CREATE TABLE IF NOT EXISTS ` + migrationsSchemaName + `.changesets_metadata (
		height INT PRIMARY KEY,
		total_chunks INT, -- total number of chunks in the changeset
		chunks_to_receive INT -- number of chunks left to receive
	)`

	// tableChangesetsSQL is the table that stores changeset chunks. These are identified by height and index.
	tableChangesetsSQL = `CREATE TABLE IF NOT EXISTS ` + migrationsSchemaName + `.changesets (
		height INT,
		index INT,
		changeset BYTEA,
		FOREIGN KEY (height) REFERENCES ` + migrationsSchemaName + `.changesets_metadata(height) ON DELETE CASCADE,
		PRIMARY KEY (height, index)
	)`

	// insertChangesetMetadataSQL is the sql query used to insert changeset metadata.
	insertChangesetMetadataSQL = `INSERT INTO ` + migrationsSchemaName + `.changesets_metadata (height, total_chunks, chunks_to_receive) VALUES ($1, $2, $3) ON CONFLICT (height) DO NOTHING;`

	// updateChangesetMetadataSQL is the sql query used to update changeset metadata.
	updateChangesetMetadataSQL = `UPDATE ` + migrationsSchemaName + `.changesets_metadata SET chunks_to_receive = chunks_to_receive - 1 WHERE height = $1;`

	// deleteChangesetMetadataSQL is the sql query used to delete changeset metadata.
	deleteChangesetMetadataSQL = `DELETE FROM ` + migrationsSchemaName + `.changesets_metadata WHERE height = $1;`

	// insertChangesetSQL is the sql query used to insert changeset.
	insertChangesetSQL = `INSERT INTO ` + migrationsSchemaName + `.changesets (height, index, changeset) VALUES ($1, $2, $3) ON CONFLICT (height, index) DO NOTHING;`

	// allChunksReceivedSQL is the sql query used to check if all chunks are received.
	allChunksReceivedSQL = `SELECT chunks_to_receive FROM ` + migrationsSchemaName + `.changesets_metadata WHERE height = $1;`

	// getChangesetsSQL is the sql query used to get changeset.
	getChangesetsSQL = `SELECT changeset FROM ` + migrationsSchemaName + `.changesets WHERE height = $1 ORDER BY index;`

	// changesetExistsSQL is the sql query used to check if a changeset exists.
	changesetExistsSQL = `SELECT 1 FROM ` + migrationsSchemaName + `.changesets WHERE height = $1 AND index = $2;`
)

// setLastChangeset sets the last changeset in the database.
func setLastChangeset(ctx context.Context, db sql.Executor, height int64) error {
	_, err := db.Execute(ctx, upsertLastChangesetSQL, defaultRowName, height)
	return err
}

// getLastChangeset gets the last changeset from the database.
func getLastChangeset(ctx context.Context, db sql.Executor) (int64, error) {
	res, err := db.Execute(ctx, getLastChangesetSQL, defaultRowName)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}

	if len(res.Rows) == 0 {
		return -1, nil
	}

	if len(res.Rows) != 1 {
		// should never happen
		return -1, fmt.Errorf("internal bug: expected one row for last changeset, got %d", len(res.Rows))
	}

	row := res.Rows[0]
	height, ok := row[0].(int64)
	if !ok {
		return -1, fmt.Errorf("internal bug: last changeset height is not an int64")
	}

	return height, nil
}

// insertChangesetMetadata inserts the changeset metadata into the database.
func insertChangesetMetadata(ctx context.Context, db sql.Executor, height int64, totalChunks int) error {
	_, err := db.Execute(ctx, insertChangesetMetadataSQL, height, totalChunks, totalChunks)
	return err
}

// deleteChangesets deletes the changesets from the database for a given height.
func deleteChangesets(ctx context.Context, db sql.Executor, height int64) error {
	// This should delete all changesets as well due to foreign key (delete on cascade) condition
	_, err := db.Execute(ctx, deleteChangesetMetadataSQL, height)
	return err
}

// insertChangesetChunk inserts a changeset chunk of index into the database.
func insertChangesetChunk(ctx context.Context, db sql.Executor, height int64, index int, changeset []byte) error {
	_, err := db.Execute(ctx, insertChangesetSQL, height, index, changeset)
	return err
}

// allChunksReceived checks if all chunks for a given height have been received.
func allChunksReceived(ctx context.Context, db sql.Executor, height int64) (bool, error) {
	res, err := db.Execute(ctx, allChunksReceivedSQL, height)
	if err != nil {
		return false, err
	}

	// row doesnt exist.
	if len(res.Rows) == 0 {
		return false, nil
	}

	if len(res.Rows) != 1 {
		// should never happen
		return false, fmt.Errorf("internal bug: expected one row for changeset metadata, got %d", len(res.Rows))
	}

	row := res.Rows[0]
	chunksToReceive, ok := row[0].(int64)
	if !ok {
		return false, fmt.Errorf("internal bug: chunks to receive is not an int64")
	}

	return chunksToReceive == 0, nil
}

// changesetChunkExists checks if a changeset chunk already exists in the database.
func changesetChunkExists(ctx context.Context, db sql.Executor, height int64, index int) (bool, error) {
	res, err := db.Execute(ctx, changesetExistsSQL, height, index)
	if err != nil {
		return false, err
	}

	return len(res.Rows) == 1, nil
}

// getChangesets gets the changesets from the database.
// It returns a byte slice of all changeset chunks in the order of chunk indexes.
func getChangesets(ctx context.Context, db sql.Executor, height int64) ([]byte, error) {
	res, err := db.Execute(ctx, getChangesetsSQL, height)
	if err != nil {
		return nil, err
	}

	var changesets []byte
	for _, row := range res.Rows {
		changeset, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("internal bug: changeset is not a byte slice")
		}

		changesets = append(changesets, changeset...)
	}

	return changesets, nil
}
