// Package meta defines a chain metadata store for the ABCI application. Prior
// to using the methods, the tables should be initialized and updated to the
// latest schema version with InitializeMetaStore.
package meta

import (
	"context"
	"fmt"
	"slices"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
)

const (
	chainSchemaName = `kwild_chain`

	chainStoreVersion = 0

	initChainTable = `CREATE TABLE IF NOT EXISTS ` + chainSchemaName + `.chain (
		height INT8 NOT NULL,
		app_hash BYTEA
	);` // no primary key, only one row

	insertChainState = `INSERT INTO ` + chainSchemaName + `.chain ` +
		`VALUES ($1, $2);`

	setChainState = `UPDATE ` + chainSchemaName + `.chain ` +
		`SET height = $1, app_hash = $2;`

	getChainState = `SELECT height, app_hash FROM ` + chainSchemaName + `.chain;`
)

func initTables(ctx context.Context, tx sql.DB) error {
	_, err := tx.Execute(ctx, initChainTable)
	return err
}

// InitializeMetaStore initializes the chain metadata store schema.
func InitializeMetaStore(ctx context.Context, db sql.DB) error {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
	}

	return versioning.Upgrade(ctx, db, chainSchemaName, upgradeFns, chainStoreVersion)
}

// GetChainState returns height and app hash from the chain state store.
// If there is no recorded data, height will be -1 and app hash nil.
func GetChainState(ctx context.Context, db sql.Executor) (int64, []byte, error) {
	res, err := db.Execute(ctx, getChainState)
	if err != nil {
		return 0, nil, err
	}

	switch n := len(res.Rows); n {
	case 0:
		return -1, nil, nil // fresh DB
	case 1:
	default:
		return 0, nil, fmt.Errorf("expected at most one row, got %d", n)
	}

	row := res.Rows[0]
	if len(row) != 2 {
		return 0, nil, fmt.Errorf("expected two columns, got %d", len(row))
	}

	height, ok := sql.Int64(row[0])
	if !ok {
		return 0, nil, fmt.Errorf("invalid type for height (%T)", res.Rows[0][0])
	}

	appHash, ok := row[1].([]byte)
	if !ok {
		return 0, nil, fmt.Errorf("expected bytes for apphash, got %T", row[1])
	}

	return height, slices.Clone(appHash), nil
}

// SetChainState will update the current height and app hash.
func SetChainState(ctx context.Context, db sql.Executor, height int64, appHash []byte) error {
	// attempt UPDATE
	res, err := db.Execute(ctx, setChainState, height, appHash)
	if err != nil {
		return err
	}
	// If no rows updated, meaning empty table, do INSERT
	if res.Status.RowsAffected == 0 {
		_, err = db.Execute(ctx, insertChainState, height, appHash)
	}
	return err
}
