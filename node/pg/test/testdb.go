package pgtest

import (
	"context"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/node/pg"
)

func NewTestPool(ctx context.Context, dropSchemas []string, dropTables ...string) (*pg.Pool, func(), error) {
	// Use pg.Pool instead of the full pg.DB since we just need a database; we
	// don't need or want the transaction or commit ID capabilities of pg.DB.
	cfg := &pg.PoolConfig{
		ConnConfig: pg.ConnConfig{
			Host:   "127.0.0.1",
			Port:   "5432",
			User:   "kwild",
			Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
			DBName: "kwil_test_db",
		},
		MaxConns: 11,
	}
	pool, err := pg.NewPool(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	dropTableFn := func() error {
		for _, drop := range dropTables {
			_, err = pool.Execute(ctx, `DROP TABLE IF EXISTS `+drop)
			if err != nil {
				return err
			}
		}
		return nil
	}
	dropSchemasFn := func() error {
		for _, drop := range dropSchemas {
			_, err = pool.Execute(ctx, `DROP SCHEMA IF EXISTS `+drop+` CASCADE`)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// if kvTableName != "" {
	// 	err = pg.CreateKVTable(ctx, kvTableName, pg.WrapQueryFun(pool.Execute))
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// _, err = pool.Execute(ctx, `TRUNCATE TABLE `+kvTableName)
	// if err != nil {
	// 	return err
	// }

	if err = dropTableFn(); err != nil {
		pool.Close()
		return nil, nil, err
	}
	if err = dropSchemasFn(); err != nil {
		pool.Close()
		return nil, nil, err
	}

	cleanUp := func() {
		dropSchemasFn()
		dropTableFn()
		pool.Close()
	}

	return pool, cleanUp, nil
}

// NewTestDB creates a new test database.
// The caller is responsible for cleaning up.
// The suggested method for cleanup is simply to have
// an outermost Tx that is rolled back at the end of the test.
func NewTestDB(t *testing.T) (db *pg.DB, err error) {
	cfg := &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
		SchemaFilter: func(s string) bool {
			return strings.Contains(s, pg.DefaultSchemaFilterPrefix)
		},
	}
	return pg.NewDB(context.Background(), cfg)
}
