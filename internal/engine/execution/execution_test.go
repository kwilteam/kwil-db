package execution_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	execution "github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/engine/types/testdata"
	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/stretchr/testify/require"
)

func Test_Execution(t *testing.T) {
	type testCase struct {
		name string
		fn   func(t *testing.T, ctx *execution.GlobalContext)
	}

	tests := []testCase{
		{
			name: "create database",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()
				err := eng.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				schema, err := eng.GetSchema(ctx, testdata.TestSchema.DBID())
				require.NoError(t, err)

				require.EqualValues(t, testdata.TestSchema, schema)
			},
		},
		{
			name: "drop database",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()
				err := eng.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = eng.DeleteDataset(ctx, testdata.TestSchema.DBID(), testdata.TestSchema.Owner)
				require.NoError(t, err)
			},
		},
		{
			name: "drop database with non-owner fails",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()
				err := eng.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = eng.DeleteDataset(ctx, testdata.TestSchema.DBID(), []byte("not_the_owner"))
				require.Error(t, err)
			},
		},
		{
			name: "drop non-existent database fails",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()

				err := eng.DeleteDataset(ctx, "not_a_real_db", testdata.TestSchema.Owner)
				require.Error(t, err)
			},
		},
		{
			name: "call an action",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()
				err := eng.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				_, err = eng.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Mutative:  true,
					Args:      []any{1, "brennan", 22},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				require.NoError(t, err)
			},
		},
		{
			name: "call an action with invalid arguments",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()

				err := eng.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				_, err = eng.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Mutative:  true,
					Args:      []any{1, "brennan"}, // missing age
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				require.Error(t, err)
			},
		},
		{
			name: "call a non-view action fails if not mutative; view action succeeds",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()

				err := eng.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				_, err = eng.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Mutative:  false,
					Args:      []any{1, "brennan", 22},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				require.Error(t, err)
				require.ErrorIs(t, err, execution.ErrMutativeProcedure)

				_, err = eng.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "get_user_by_address",
					Mutative:  false,
					Args:      []any{"address"},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				require.NoError(t, err)
			},
		},
		{
			name: "call an extension",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()

				err := eng.CreateDataset(ctx, testSchema, testSchema.Owner)
				require.NoError(t, err)

				_, err = eng.Execute(ctx, &types.ExecutionData{
					Dataset:   testSchema.DBID(),
					Procedure: "use_math",
					Mutative:  true,
					Args:      []any{1, 2},
					Signer:    testSchema.Owner,
					Caller:    string(testSchema.Owner),
				})
				require.NoError(t, err)

				// call non-mutative
				// since we do not have a sql connection, we cannot evaluate the result
				_, err = eng.Execute(ctx, &types.ExecutionData{
					Dataset:   testSchema.DBID(),
					Procedure: "use_math",
					Mutative:  false,
					Args:      []any{1, 2},
					Signer:    testSchema.Owner,
					Caller:    string(testSchema.Owner),
				})
				require.NoError(t, err)
			},
		},
		{
			name: "list datasets",
			fn: func(t *testing.T, eng *execution.GlobalContext) {
				ctx := context.Background()

				owner := "owner"

				err := eng.CreateDataset(ctx, testdata.TestSchema, []byte(owner))
				require.NoError(t, err)

				datasets, err := eng.ListDatasets(ctx, []byte(owner))
				require.NoError(t, err)

				require.Equal(t, 1, len(datasets))
				require.Equal(t, testdata.TestSchema.Name, datasets[0].Name)
				require.Equal(t, testdata.TestSchema.Owner, datasets[0].Owner)
				require.Equal(t, testdata.TestSchema.DBID(), datasets[0].DBID)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mth := &mathInitializer{}

			ctx := context.Background()
			engine, err := execution.NewGlobalContext(ctx,
				&mockRegistry{
					dbs: map[string]*trackedDB{},
				}, map[string]execution.NamespaceInitializer{
					"math": mth.initialize,
				},
			)
			require.NoError(t, err)
			tc.fn(t, engine)
		})
	}
}

type trackedDB struct {
	kv map[string][]byte
}

// mockRegistry is a mock database registry
type mockRegistry struct {
	dbs map[string]*trackedDB
}

func (m *mockRegistry) Create(ctx context.Context, dbid string) error {
	_, ok := m.dbs[dbid]
	if ok {
		return fmt.Errorf(`database "%s" already exists`, dbid)
	}

	m.dbs[dbid] = &trackedDB{
		kv: make(map[string][]byte),
	}

	return nil
}

func (m *mockRegistry) Delete(ctx context.Context, dbid string) error {
	_, ok := m.dbs[dbid]
	if !ok {
		return fmt.Errorf(`database "%s" does not exist`, dbid)
	}

	delete(m.dbs, dbid)
	return nil
}

func (m *mockRegistry) Execute(ctx context.Context, dbid string, stmt string, params map[string]any) (*sql.ResultSet, error) {
	_, ok := m.dbs[dbid]
	if !ok {
		return nil, fmt.Errorf(`database "%s" does not exist`, dbid)
	}

	return &sql.ResultSet{
		ReturnedColumns: []string{},
		Rows:            [][]any{},
	}, nil
}

func (m *mockRegistry) List(ctx context.Context) ([]string, error) {
	dbs := make([]string, 0)
	for dbid := range m.dbs {
		dbs = append(dbs, dbid)
	}

	return dbs, nil
}

func (m *mockRegistry) Query(ctx context.Context, dbid string, stmt string, params map[string]any) (*sql.ResultSet, error) {
	_, ok := m.dbs[dbid]
	if !ok {
		return nil, fmt.Errorf(`database "%s" does not exist`, dbid)
	}

	return &sql.ResultSet{
		ReturnedColumns: []string{},
		Rows:            [][]any{},
	}, nil
}

func (m *mockRegistry) Set(ctx context.Context, dbid string, key, value []byte) error {
	db, ok := m.dbs[dbid]
	if !ok {
		return fmt.Errorf(`database "%s" does not exist`, dbid)
	}

	db.kv[string(key)] = value
	return nil
}

func (m *mockRegistry) Get(ctx context.Context, dbid string, key []byte, sync bool) ([]byte, error) {
	db, ok := m.dbs[dbid]
	if !ok {
		return nil, fmt.Errorf(`database "%s" does not exist`, dbid)
	}

	key, ok = db.kv[string(key)]
	if !ok {
		return nil, nil
	}

	return key, nil
}

// identitySchema is a schema that relies on the testdata user's schema
// it creates an example credential application
var testSchema = &types.Schema{
	Name:  "identity_db",
	Owner: []byte(`owner`),
	Tables: []*types.Table{
		{
			Name: "credentials",
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
					Name: "user_id",
					Type: types.INT,
					Attributes: []*types.Attribute{
						{
							Type: types.NOT_NULL,
						},
					},
				},
				{
					Name: "credential",
					Type: types.TEXT,
				},
			},
			Indexes: []*types.Index{
				{
					Name:    "user_id",
					Columns: []string{"user_id"},
					Type:    types.BTREE,
				},
			},
		},
	},
	Procedures: []*types.Procedure{
		{
			Name:   "use_math",
			Args:   []string{"$a", "$b"},
			Public: true,
			Modifiers: []types.Modifier{
				types.ModifierView,
			},
			Statements: []string{
				`math.add($a, $b);`,
			},
		},
	},
	Extensions: []*types.Extension{
		{
			Name: "math",
			Initialization: []*types.ExtensionConfig{
				{
					Key:   "math_key",
					Value: "math_val",
				},
			},
			Alias: "math",
		},
	},
}

// mocks a namespace initializer
type mathInitializer struct {
	vals map[string]string
}

func (m *mathInitializer) initialize(_ context.Context, mp map[string]string) (execution.Namespace, error) {
	m.vals = mp

	return &mathExt{}, nil
}

type mathExt struct{}

var _ execution.Namespace = &mathExt{}

func (m *mathExt) Call(caller *execution.ScopeContext, method string, inputs []any) ([]any, error) {
	return nil, nil
}

// Test_OrderSchemas tests that schemas are ordered correctly when importing with dependencies
func Test_OrderSchemas(t *testing.T) {
	// create random schemas, and randomly add others as dependencies
	schemas := make([]*types.Schema, 0)

	for i := 0; i < 100; i++ {
		schema := randomSchema()

		for _, schema2 := range schemas {
			schema2.Extensions = append(schema.Extensions, &types.Extension{
				Name:  schema.DBID(),
				Alias: schema.Name,
			})
		}

		schemas = append(schemas, schema)
	}

	// add some more that have zero dependencies
	for _, schema := range schemas {
		for i := 0; i < 10; i++ {
			dep := randomSchema()
			schema.Extensions = append(schema.Extensions, &types.Extension{
				Name:  dep.DBID(),
				Alias: dep.Name,
			})

			// add the dependency to the list of schemas
			schemas = append(schemas, dep)
		}
	}

	// now create a datastore to see if it imports the schemas in the correct order

	ctx := context.Background()
	mth := &mathInitializer{}
	_, err := execution.NewGlobalContext(ctx,
		&mockRegistry{
			dbs: make(map[string]*trackedDB),
		}, map[string]execution.NamespaceInitializer{
			"math": mth.initialize,
		},
	)
	require.NoError(t, err)

}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomSchema() *types.Schema {
	return &types.Schema{
		Name:  randomString(10),
		Owner: []byte(randomString(10)),
	}
}
