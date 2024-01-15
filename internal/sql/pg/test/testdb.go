package test

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

func NewTestPool(ctx context.Context, dropSchemas []string, dropTables ...string) (*pg.Pool, func(), error) {
	// Use pg.Pool instead of the full pg.DB since we just need a database; we
	// don't need or want the transaction or commit ID capabilities of pg.DB.
	cfg := &pg.PoolConfig{
		ConnConfig: pg.ConnConfig{
			Host:   "/var/run/postgresql",
			Port:   "",
			User:   "kwil_test_user",
			Pass:   "kwil", // would be ignored if pg_hba.conf set with trust
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
