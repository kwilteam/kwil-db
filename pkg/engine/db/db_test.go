package db_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/db/test"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
	"github.com/stretchr/testify/assert"
)

var (
	testExt1 = &types.Extension{
		Name: "testExt1",
		Initialization: map[string]string{
			"testConfig1": "testValue1",
		},
		Alias: "testAlias1",
	}

	testExt2 = &types.Extension{
		Name: "testExt2",
		Initialization: map[string]string{
			"testConfig2": "testValue2",
		},
		Alias: "testAlias2",
	}
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
				err := datastore.CreateTable(ctx, &testdata.Table_users)
				assert.NoError(t, err)

				err = datastore.CreateTable(ctx, &testdata.Table_posts)
				assert.NoError(t, err)

				tbls, err := datastore.ListTables(ctx)
				assert.NoError(t, err)
				assert.Len(t, tbls, 2)

				containsUsers := false
				containsPosts := false
				for _, tbl := range tbls {
					if reflect.DeepEqual(tbl, &testdata.Table_users) {
						containsUsers = true
					}

					if reflect.DeepEqual(tbl, &testdata.Table_posts) {
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

				err := datastore.StoreProcedure(ctx, &testdata.Procedure_create_user)
				assert.NoError(t, err)

				err = datastore.StoreProcedure(ctx, &testdata.Procedure_create_post)
				assert.NoError(t, err)

				procs, err := datastore.ListProcedures(ctx)
				assert.NoError(t, err)
				assert.Len(t, procs, 2)

				containsGetUser := false
				containsGetPost := false
				for _, proc := range procs {
					if reflect.DeepEqual(proc, &testdata.Procedure_create_user) {
						containsGetUser = true
					}

					if reflect.DeepEqual(proc, &testdata.Procedure_create_post) {
						containsGetPost = true
					}
				}

				assert.True(t, containsGetUser)
				assert.True(t, containsGetPost)
			},
		},
		{
			name: "create 2 extensions and retrieve them",
			test: func(t *testing.T, datastore *db.DB) {

				ctx := context.Background()

				err := datastore.StoreExtension(ctx, testExt1)
				assert.NoError(t, err)

				err = datastore.StoreExtension(ctx, testExt2)
				assert.NoError(t, err)

				exts, err := datastore.ListExtensions(ctx)
				assert.NoError(t, err)
				assert.Len(t, exts, 2)

				containsExt1 := false
				containsExt2 := false
				for _, ext := range exts {
					if reflect.DeepEqual(ext, testExt1) {
						containsExt1 = true
					}

					if reflect.DeepEqual(ext, testExt2) {
						containsExt2 = true
					}
				}

				assert.True(t, containsExt1)
				assert.True(t, containsExt2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			datastore, teardown, err := test.OpenTestDB(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer datastore.Close()
			defer teardown()

			tt.test(t, datastore)
		})
	}
}

var defaultTables = []*types.Table{
	&testdata.Table_users,
	&testdata.Table_posts,
}

func Test_Prepare(t *testing.T) {
	type testCase struct {
		name      string
		tables    []*types.Table
		statement string
		wantErr   bool
	}

	tests := []testCase{
		{
			name:      "valid SELECT statement",
			tables:    defaultTables,
			statement: `SELECT * FROM users WHERE id = $id;`,
			wantErr:   false,
		},
		{
			name:      "invalid group by",
			tables:    defaultTables,
			statement: `SELECT * FROM users WHERE id = $id GROUP BY id;`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			datastore, teardown, err := test.OpenTestDB(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer datastore.Close()
			defer teardown()

			for _, tbl := range tt.tables {
				err := datastore.CreateTable(ctx, tbl)
				assert.NoError(t, err)
			}

			_, err = datastore.Prepare(ctx, tt.statement)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
