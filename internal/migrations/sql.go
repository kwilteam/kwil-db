package migrations

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common/sql"
)

// InitializeMigrationSchema initializes the migration schema in the database.
func InitializeMigrationSchema(ctx context.Context, db sql.DB) error {
	_, err := db.Execute(ctx, tableMigrationsSQL)
	return err
}

const migrationsSchemaName = `kwild_migrations`

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
	// migrationIsActiveSQL is the sql query used to check if a migration is active.
	migrationIsActiveSQL = `SELECT COUNT(*) FROM ` + migrationsSchemaName + `.migration;`
	// createMigrationSQL is the sql query used to create a new migration.
	createMigrationSQL = `INSERT INTO ` + migrationsSchemaName + `.migration (id, start_height, end_height, chain_id) VALUES ($1, $2, $3, $4);`
)

// getMigrationState gets the current migration state from the database.
func getMigrationState(ctx context.Context, db sql.Executor) (*activeMigration, error) {
	res, err := db.Execute(ctx, getMigrationSQL)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("network does not have an active migration")
	}
	if err != nil {
		return nil, err
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
	count, ok := row[0].(int64)
	if !ok {
		return false, fmt.Errorf("internal bug: count is not an int64")
	}

	return count == 1, nil
}

// createMigration creates a new migration state in the database.
func createMigration(ctx context.Context, db sql.Executor, md *activeMigration) error {
	_, err := db.Execute(ctx, createMigrationSQL, 1, md.StartHeight, md.EndHeight, md.ChainID)
	return err
}
