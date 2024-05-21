package execution

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/common"
	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/testdata"
	"github.com/kwilteam/kwil-db/extensions/precompiles"

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

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				schema, err := eng.GetSchema(testdata.TestSchema.DBID())
				assert.NoError(t, err)

				assert.EqualValues(t, testdata.TestSchema, schema)
			},
		},
		{
			name: "drop database",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				_, ok := db.dbs[testdata.TestSchema.DBID()]
				assert.True(t, ok)

				err = eng.DeleteDataset(ctx, db, testdata.TestSchema.DBID(), &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid2",
				})
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

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				err = eng.DeleteDataset(ctx, db, testdata.TestSchema.DBID(), &common.TransactionData{
					Signer: []byte("not_owner"),
					Caller: "not_owner",
					TxID:   "txid1",
				})
				assert.Error(t, err)
			},
		},
		{
			name: "drop non-existent database fails",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.DeleteDataset(ctx, db, "not_a_real_db", &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid2",
				})
				assert.Error(t, err)
			},
		},
		{
			name: "call an action",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				_, err = eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Args:      []any{1, "brennan", 22},
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid2",
					},
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "call an action with invalid arguments",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				_, err = eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Args:      []any{1, "brennan"}, // missing age
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid2",
					},
				})
				assert.Error(t, err)
			},
		},
		{
			name: "call a recursive procedure",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				_, err = eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionRecursive.Name,
					Args:      []any{"id000000", "asdfasdfasdfasdf", "bigbigbigbigbigbigbigbigbigbig"},
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid2",
					},
				})
				assert.ErrorIs(t, err, ErrMaxStackDepth)
			},
		},
		{
			name: "call a procedure that hits max call stack depth less directly",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				_, err = eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionRecursiveSneakyA.Name,
					Args:      []any{},
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid2",
					},
				})
				assert.ErrorIs(t, err, ErrMaxStackDepth)
			},
		},
		{
			name: "call a non-view action fails if not mutative; view action succeeds",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				db2 := newDB(true)

				_, err = eng.Procedure(ctx, db2, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "create_user",
					Args:      []any{1, "brennan", 22},
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid2",
					},
				})
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrMutativeProcedure)

				_, err = eng.Procedure(ctx, db2, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: "get_user_by_address",
					Args:      []any{"address"},
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid3",
					},
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "call an extension",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := newDB(false)

				err := eng.CreateDataset(ctx, db, testSchema, &common.TransactionData{
					Signer: testSchema.Owner,
					Caller: string(testSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				_, err = eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testSchema.DBID(),
					Procedure: "use_math",
					Args:      []any{1, 2},
					TransactionData: common.TransactionData{
						Signer: testSchema.Owner,
						Caller: string(testSchema.Owner),
						// no txid since it is non-mutative
					},
				})
				assert.NoError(t, err)

				// call non-mutative
				// since we do not have a sql connection, we cannot evaluate the result
				_, err = eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testSchema.DBID(),
					Procedure: "use_math",
					Args:      []any{1, 2},
					TransactionData: common.TransactionData{
						Signer: testSchema.Owner,
						Caller: string(testSchema.Owner),
						TxID:   "txid3",
					},
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

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: []byte(owner),
					Caller: owner,
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				datasets, err := eng.ListDatasets([]byte(owner))
				assert.NoError(t, err)

				assert.Equal(t, 1, len(datasets))
				assert.Equal(t, testdata.TestSchema.Name, datasets[0].Name)
				assert.Equal(t, testdata.TestSchema.Owner, datasets[0].Owner)
				assert.Equal(t, testdata.TestSchema.DBID(), datasets[0].DBID)
			},
		},
		{
			name: "procedure returning table",
			fn: func(t *testing.T, eng *GlobalContext) {
				ctx := context.Background()
				db := mockResultDB(&sql.ResultSet{
					Columns: []string{"_out_id", "_out_name", "_out_age"},
				})

				owner := "owner"

				err := eng.CreateDataset(ctx, db, testdata.TestSchema, &common.TransactionData{
					Signer: []byte(owner),
					Caller: owner,
					TxID:   "txid1",
				})
				assert.NoError(t, err)

				res, err := eng.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcGetUsersByAge.Name,
					Args:      []any{22},
					TransactionData: common.TransactionData{
						Signer: []byte(owner),
						Caller: owner,
					},
				})
				assert.NoError(t, err)

				for i, expected := range testdata.ProcGetUsersByAge.Returns.Fields {
					assert.Equal(t, expected.Name, res.Columns[i])
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mth := &mathInitializer{}

			ctx := context.Background()

			engine, err := NewGlobalContext(ctx,
				newDB(false), map[string]precompiles.Initializer{
					"math": mth.initialize,
				}, &common.Service{
					Logger:           log.NewNoOp().Sugar(),
					ExtensionConfigs: map[string]map[string]string{},
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

// mockResultDB can be used to mock a result set for a query
func mockResultDB(result *sql.ResultSet) *mockDB {
	db := newDB(false)
	db.resultSet = result

	return db
}

type mockDB struct {
	accessMode    sql.AccessMode
	dbs           map[string][]byte // serialized schemas
	executedStmts []string
	resultSet     *sql.ResultSet
}

var _ sql.AccessModer = (*mockDB)(nil)

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
		// first arg is uuid, 2nd is dbid, 3rd is schema content, 4th is schema version
		m.dbs[args[1].(string)] = args[2].([]byte)
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

		if m.resultSet != nil {
			return m.resultSet, nil
		}
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
					Type: types.IntType,
					Attributes: []*types.Attribute{
						{
							Type: types.PRIMARY_KEY,
						},
					},
				},
				{
					Name: "user_id",
					Type: types.IntType,
					Attributes: []*types.Attribute{
						{
							Type: types.NOT_NULL,
						},
					},
				},
				{
					Name: "credential",
					Type: types.TextType,
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
	Actions: []*types.Action{
		{
			Name:       "use_math",
			Parameters: []string{"$a", "$b"},
			Public:     true,
			Modifiers: []types.Modifier{
				types.ModifierView,
			},
			Body: `math.add($a, $b);`,
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

func (m *mathInitializer) initialize(_ *precompiles.DeploymentContext, _ *common.Service, mp map[string]string) (precompiles.Instance, error) {
	m.vals = mp

	return &mathExt{}, nil
}

type mathExt struct{}

var _ precompiles.Instance = &mathExt{}

func (m *mathExt) Call(caller *precompiles.ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
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
		newDB(false), map[string]precompiles.Initializer{
			"math": mth.initialize,
		}, &common.Service{
			Logger:           log.NewNoOp().Sugar(),
			ExtensionConfigs: map[string]map[string]string{},
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
