package execution

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/engine/types/testdata"
	sql "github.com/kwilteam/kwil-db/internal/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Execution(t *testing.T) {
	type testCase struct {
		name string
		fn   func(t *testing.T, ctx *GlobalContext)
	}

	tests := []testCase{
		{
			name: "create database",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)

				schema, err := eng.GetSchema(ctx, testdata.TestSchema.DBID())
				assert.NoError(t, err)

				assert.EqualValues(t, testdata.TestSchema, schema)
			},
		},
		{
			name: "drop database",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)

				_, ok := db.dbs[testdata.TestSchema.DBID()]
				assert.True(t, ok)

				err = eng.DeleteDataset(ctx, db, testdata.TestSchema.DBID(), testdata.TestSchema.Owner)
				assert.NoError(t, err)

				_, ok = db.dbs[testdata.TestSchema.DBID()]
				assert.False(t, ok)
			},
		},
		{
			name: "drop database with non-owner fails",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)

				err = eng.DeleteDataset(ctx, db, testdata.TestSchema.DBID(), []byte("not_the_owner"))
				assert.Error(t, err)
			},
		},
		{
			name: "drop non-existent database fails",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.DeleteDataset(ctx, db, "not_a_real_db", testdata.TestSchema.Owner)
				assert.Error(t, err)
			},
		},
		{
			name: "call an action",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)

				_, err = eng.Execute(ctx, db, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Args:      []any{1, "brennan", 22},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "call an action with invalid arguments",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)

				_, err = eng.Execute(ctx, db, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Args:      []any{1, "brennan"}, // missing age
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				assert.Error(t, err)
			},
		},
		{
			name: "call a non-view action fails if not mutative; view action succeeds",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)

				db2 := newDB(true)

				_, err = eng.Execute(ctx, db2, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Args:      []any{1, "brennan", 22},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrMutativeProcedure)

				_, err = eng.Execute(ctx, db2, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "get_user_by_address",
					Args:      []any{"address"},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "call an extension",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testSchema, testSchema.Owner)
				assert.NoError(t, err)

				_, err = eng.Execute(ctx, db, &types.ExecutionData{
					Dataset:   testSchema.DBID(),
					Procedure: "use_math",
					Args:      []any{1, 2},
					Signer:    testSchema.Owner,
					Caller:    string(testSchema.Owner),
				})
				assert.NoError(t, err)

				// call non-mutative
				// since we do not have a sql connection, we cannot evaluate the result
				_, err = eng.Execute(ctx, db, &types.ExecutionData{
					Dataset:   testSchema.DBID(),
					Procedure: "use_math",
					Args:      []any{1, 2},
					Signer:    testSchema.Owner,
					Caller:    string(testSchema.Owner),
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "list datasets",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				owner := "owner"

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, []byte(owner))
				assert.NoError(t, err)

				datasets, err := eng.ListDatasets(ctx, []byte(owner))
				assert.NoError(t, err)

				assert.Equal(t, 1, len(datasets))
				assert.Equal(t, testdata.TestSchema.Name, datasets[0].Name)
				assert.Equal(t, testdata.TestSchema.Owner, datasets[0].Owner)
				assert.Equal(t, testdata.TestSchema.DBID(), datasets[0].DBID)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mth := &mathInitializer{}

			ctx := context.Background()
			engine, err := NewGlobalContext(ctx,
				newDB(false), map[string]ExtensionInitializer{
					"math": mth.initialize,
				},
			)
			require.NoError(t, err)
			tc.fn(t, engine)
		})
	}
}

func newDB(readonly bool) *mockDB {
	am := sql.ReadWrite
	if readonly {
		am = sql.ReadOnly
	}

	return &mockDB{
		accessMode:    am,
		dbs:           make(map[string][]byte),
		executedStmts: make([]string, 0),
	}
}

type mockDB struct {
	accessMode    sql.AccessMode
	dbs           map[string][]byte // serialized schemas
	executedStmts []string
}

func (m *mockDB) AccessMode() sql.AccessMode {
	return m.accessMode
}

func (m *mockDB) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{
		m,
	}, nil
}

func (m *mockDB) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	// mock some expected queries used internally
	switch stmt {
	case sqlStoreKwilSchema:
		// first arg is dbid, 2nd is schema content, 3rd is schema version
		m.dbs[args[0].(string)] = args[1].([]byte)
	case sqlListSchemaContent:
		rows := make([][]any, 0)
		for _, bts := range m.dbs {
			rows = append(rows, []any{bts})
		}

		return &sql.ResultSet{
			Columns: []string{"schema_content"},
			Rows:    rows,
		}, nil
	case sqlDeleteKwilSchema:
		delete(m.dbs, args[0].(string))
	default:
		m.executedStmts = append(m.executedStmts, stmt)
	}

	return &sql.ResultSet{
		Columns: []string{},
		Rows:    [][]any{},
	}, nil
}

type mockTx struct {
	*mockDB
}

func (m *mockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return nil
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

func (m *mathInitializer) initialize(_ *DeploymentContext, mp map[string]string) (ExtensionNamespace, error) {
	m.vals = mp

	return &mathExt{}, nil
}

type mathExt struct{}

var _ ExtensionNamespace = &mathExt{}

func (m *mathExt) Call(caller *ProcedureContext, method string, inputs []any) ([]any, error) {
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
	_, err := NewGlobalContext(ctx,
		newDB(false), map[string]ExtensionInitializer{
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
