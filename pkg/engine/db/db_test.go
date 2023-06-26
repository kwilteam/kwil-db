package db_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/types/tsts"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
	"github.com/stretchr/testify/assert"
)

type testFunc func(*testing.T, *db.DB)

func Test_CreateTables(t *testing.T) {
	tests := []struct {
		name string
		test testFunc
	}{
		{
			name: "create 2 tables",
			test: func(t *testing.T, datastore *db.DB) {
				ctx := context.Background()
				err := datastore.CreateTable(ctx, &tsts.Table_users)
				assert.NoError(t, err)

				err = datastore.CreateTable(ctx, &tsts.Table_posts)
				assert.NoError(t, err)

				tbls, err := datastore.ListTables(ctx)
				assert.NoError(t, err)
				assert.Len(t, tbls, 2)

				containsUsers := false
				containsPosts := false
				for _, tbl := range tbls {
					if reflect.DeepEqual(tbl, &tsts.Table_users) {
						containsUsers = true
					}

					if reflect.DeepEqual(tbl, &tsts.Table_posts) {
						containsPosts = true
					}
				}

				assert.True(t, containsUsers)
				assert.True(t, containsPosts)
			},
		},
		{
			name: "create 2 procedures and retrieve them",
			test: func(t *testing.T, datastore *db.DB) {

				ctx := context.Background()

				err := datastore.StoreProcedure(ctx, &tsts.Procedure_create_user)
				assert.NoError(t, err)

				err = datastore.StoreProcedure(ctx, &tsts.Procedure_create_post)
				assert.NoError(t, err)

				procs, err := datastore.ListProcedures(ctx)
				assert.NoError(t, err)
				assert.Len(t, procs, 2)

				containsGetUser := false
				containsGetPost := false
				for _, proc := range procs {
					if reflect.DeepEqual(proc, &tsts.Procedure_create_user) {
						containsGetUser = true
					}

					if reflect.DeepEqual(proc, &tsts.Procedure_create_post) {
						containsGetPost = true
					}
				}

				assert.True(t, containsGetUser)
				assert.True(t, containsGetPost)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			testDb, td, err := sqlTesting.OpenTestDB()
			if err != nil {
				t.Fatal(err)
			}
			defer td()

			datastore, err := db.NewDB(ctx, databaseAdapter{testDb})
			if err != nil {
				t.Fatal(err)
			}
			defer datastore.Close()

			tt.test(t, datastore)
		})
	}
}

type databaseAdapter struct {
	sqlTesting.TestSqliteClient
}

func (d databaseAdapter) Prepare(query string) (db.Statement, error) {
	return d.TestSqliteClient.Prepare(query)
}

func (d databaseAdapter) Savepoint() (db.Savepoint, error) {
	return d.TestSqliteClient.Savepoint()
}
