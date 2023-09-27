package db_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/db/test"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql"
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
				psts := tblPosts
				_ = psts
				usrs := tblUsers
				_ = usrs

				ctx := context.Background()
				err := datastore.CreateTable(ctx, tblUsers)
				assert.NoError(t, err)

				err = datastore.CreateTable(ctx, tblPosts)
				assert.NoError(t, err)

				tbls, err := datastore.ListTables(ctx)
				assert.NoError(t, err)
				assert.Len(t, tbls, 2)

				containsUsers := false
				containsPosts := false
				for _, tbl := range tbls {

					if deepEqual(tbl, tblUsers) {
						containsUsers = true
					}

					if deepEqual(tbl, tblPosts) {
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

				err := datastore.StoreProcedure(ctx, procedureCreateUser)
				assert.NoError(t, err)

				err = datastore.StoreProcedure(ctx, procedureCreatePost)
				assert.NoError(t, err)

				procs, err := datastore.ListProcedures(ctx)
				assert.NoError(t, err)
				assert.Len(t, procs, 2)

				containsGetUser := false
				containsGetPost := false
				for _, proc := range procs {
					if deepEqual(proc, procedureCreateUser) {
						containsGetUser = true
					}

					if deepEqual(proc, procedureCreatePost) {
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
					if deepEqual(ext, testExt1) {
						containsExt1 = true
					}

					if deepEqual(ext, testExt2) {
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

// deepEqual does a deep comparison, while considering empty slices as equal to nils.
func deepEqual(a, b any) bool {
	return cmp.Equal(a, b, cmpopts.EquateEmpty())
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
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type baseMockDatastore struct {
}

func (b *baseMockDatastore) Close() error {
	return nil
}

func (b *baseMockDatastore) Delete() error {
	return nil
}

func (b *baseMockDatastore) Execute(ctx context.Context, stmt string, args map[string]any) error {
	return nil
}

func (b *baseMockDatastore) Prepare(stmt string) (sql.Statement, error) {
	return nil, nil
}

func (b *baseMockDatastore) Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error) {
	return nil, nil
}

func (b *baseMockDatastore) Savepoint() (sql.Savepoint, error) {
	return nil, nil
}

func (b *baseMockDatastore) TableExists(ctx context.Context, table string) (bool, error) {
	return false, nil
}

var (
	tblUsers = &types.Table{
		Name: "users",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.INT,
				Attributes: []*types.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "name",
				Type: types.TEXT,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MIN_LENGTH,
						Value: "3",
					},
					{
						Type:  types.MAX_LENGTH,
						Value: "255",
					},
					{
						Type: types.UNIQUE,
					},
				},
			},
			{
				Name: "age",
				Type: types.INT,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MIN,
						Value: "0",
					},
					{
						Type:  types.MAX,
						Value: "150",
					},
				},
			},
		},
		Indexes: []*types.Index{
			{
				Name: "age_index",
				Columns: []string{
					"age",
				},
				Type: types.BTREE,
			},
		},
	}

	tblPosts = &types.Table{
		Name: "posts",
		Columns: []*types.Column{
			{
				Name: "id1",
				Type: types.INT,
			},
			{
				Name: "id2", // doing this to check composite primary keys
				Type: types.INT,
			},
			{
				Name: "title",
				Type: types.TEXT,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
			{
				Name: "author_id",
				Type: types.INT,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
		},
		Indexes: []*types.Index{
			{
				Name: "primary_key",
				Columns: []string{
					"id1",
					"id2",
				},
				Type: types.PRIMARY,
			},
		},
		ForeignKeys: []*types.ForeignKey{
			{
				ChildKeys:   []string{"author_id"},
				ParentKeys:  []string{"id"},
				ParentTable: "users",
				Actions: []*types.ForeignKeyAction{
					{
						On: types.ON_UPDATE,
						Do: types.DO_CASCADE,
					},
				},
			},
		},
	}
)

var defaultTables = []*types.Table{
	tblUsers,
	tblPosts,
}

var (
	procedureCreateUser = &types.Procedure{
		Name:      "create_user",
		Args:      []string{"$id", "$name", "$age"},
		Public:    false,
		Modifiers: []types.Modifier{types.ModifierAuthenticated},
		Statements: []string{
			"INSERT INTO users (id, name, age) VALUES ($id, $name, $age);",
		},
	}

	procedureCreatePost = &types.Procedure{
		Name:   "create_post",
		Args:   []string{"$id1", "$id2", "$title", "$author_id"},
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierAuthenticated,
			types.ModifierOwner,
		},
		Statements: []string{
			"INSERT INTO posts (id1, id2, title, author_id) VALUES ($id1, $id2, $title, $author_id);",
		},
	}
)
