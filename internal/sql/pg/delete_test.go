package pg_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

// TODO: delete this file

func Test_Del(t *testing.T) {
	ctx := context.Background()
	db, err := pg.NewDB(ctx, &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			MaxConns: 2,
			ConnConfig: pg.ConnConfig{
				Host:   "localhost",
				Port:   "5432",
				User:   "postgres",
				Pass:   "password",
				DBName: "postgres",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.AutoCommit(true)

	res, err := db.Execute(ctx, "SELECT $1 + $2;", pg.QueryModeInferredArgTypes, int64(1), int64(1))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}
