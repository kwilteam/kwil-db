package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	engine "github.com/kwilteam/kwil-db/pkg/engine"
	engineTesting "github.com/kwilteam/kwil-db/pkg/engine/testing"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/stretchr/testify/assert"
)

const testPrivateKey = "4a3142b366011d28c2a3ca33a678ff753c978c685178d4168bad4474ea480cc9"

func newTestUser() *types.User {
	pk, err := crypto.Secp256k1PrivateKeyFromHex(testPrivateKey)
	if err != nil {
		panic(err)
	}

	return &types.User{
		PublicKey: pk.PubKey().Bytes(),
		AuthType:  auth.EthPersonalSignAuth,
	}
}

var (
	testTables = []*types.Table{
		testdata.TableUsers,
		testdata.TablePosts,
	}

	testProcedures = []*types.Procedure{
		procedureCreateUser,
		procedureCreatePost,
	}

	testInitializedExtensions = []*types.Extension{
		{
			Name: "erc20",
			Initialization: map[string]string{
				"address": "0x1234",
			},
			Alias: "usdc",
		},
	}
)

func Test_Open(t *testing.T) {
	ctx := context.Background()
	user := newTestUser()

	e, teardown, err := engineTesting.NewTestEngine(ctx, newMockRegister(),
		engine.WithExtensions(testExtensions),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	_, err = e.CreateDataset(ctx, &types.Schema{
		Name:       "testdb1",
		Extensions: testInitializedExtensions,
		Tables:     testTables,
		Procedures: testProcedures,
	}, user)
	if err != nil {
		t.Fatal(err)
	}

	// close the engine
	// we likely need some more tests regarding this, as well as orphaned records.
	err = e.Close()
	if err != nil {
		t.Fatal(err)
	}

	e2, teardown2, err := engineTesting.NewTestEngine(ctx, newMockRegister(),
		engine.WithExtensions(testExtensions),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown2()

	// check if the dataset was created
	dataset, err := e2.GetDataset(ctx, utils.GenerateDBID("testdb1", user.PublicKey))
	if err != nil {
		t.Fatal(err)
	}

	// check if the dataset has the correct tables
	tables, err := dataset.ListTables(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, table := range tables {
		if !deepEqual(table, findTable(table.Name)) {
			t.Errorf("tables not equal: %v, %v", table, findTable(table.Name))
		}
	}

	// check if the dataset has the correct procedures
	procs := dataset.ListProcedures()

	for _, proc := range procs {
		if !deepEqual(proc, findProc(proc.Name)) {
			t.Errorf("procedures not equal: %v, %v", proc, findProc(proc.Name))
		}
	}

	// list the datasets
	datasets, err := e2.ListDatasets(ctx, user.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	assert.ElementsMatch(t, []string{"testdb1"}, datasets)
}

func findProc(name string) *types.Procedure {
	for _, proc := range testProcedures {
		if proc.Name == name {
			return proc
		}
	}

	panic("procedure not found")
}

func findTable(name string) *types.Table {
	for _, table := range testTables {
		if table.Name == name {
			return table
		}
	}

	panic("table not found")
}

func Test_CreateDataset(t *testing.T) {
	type execution struct {
		procedure string
		args      []any
	}

	tests := []struct {
		name     string
		database *types.Schema
		exec     []*execution
		wantErr  bool
	}{
		{
			name: "create a dataset with a variety of statements",
			database: &types.Schema{
				Name:       "test_db",
				Extensions: testInitializedExtensions,
				Tables:     testTables,
				Procedures: []*types.Procedure{
					{
						Name:   "create_user",
						Args:   []string{"$id", "$username", "$age"},
						Public: true,
						Statements: []string{
							"$current_bal = usdc.balanceOf(@caller);",
							"SELECT case when $current_bal < 100 then ERROR('not enough balance') end;",
							"INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "create a dataset with invalid dml",
			database: &types.Schema{
				Name:       "test_db",
				Extensions: testInitializedExtensions,
				Tables:     testTables,
				Procedures: []*types.Procedure{
					{
						Name:   "create_user",
						Args:   []string{"$id", "$username", "$age"},
						Public: true,
						Statements: []string{
							"INSERT INTO the table users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ensure procedures, tables, columns, extensions are case insensitive",
			database: &types.Schema{
				Name: "test_db",
				Extensions: []*types.Extension{
					{
						Name: "eRC20",
						Initialization: map[string]string{
							"address": "0x1234", // initializations should not be case insensitive
						},
						Alias: "usDc",
					},
				},
				Tables: []*types.Table{
					{
						Name: "usERs",
						Columns: []*types.Column{
							{
								Name: "iD",
								Type: types.INT,
								Attributes: []*types.Attribute{
									{
										Type: types.PRIMARY_KEY,
									},
								},
							},
							{
								Name: "useRName",
								Type: types.TEXT,
							},
						},
					},
				},
				Procedures: []*types.Procedure{
					{
						Name:   "creAte_User",
						Args:   []string{"$id", "$username"},
						Public: true,
						Statements: []string{
							"$cuRRent_bal = uSdc.balanceOf(@caller);",
							"INSERT INTO Users (id, uSername) VALUES ($id, $username);",
						},
					},
				},
			},
			exec: []*execution{
				{
					procedure: "create_user",
					args:      []any{1, "test"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			e, teardown, err := engineTesting.NewTestEngine(ctx, newMockRegister(),
				engine.WithExtensions(testExtensions),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer teardown()

			_, err = e.CreateDataset(ctx, tt.database, newTestUser())
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("Engine.CreateDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hasErr {
				return
			}

			// check if the dataset was created
			_, err = e.GetDataset(ctx, utils.GenerateDBID(tt.database.Name, tt.database.Owner))
			assert.NoError(t, err)

			for _, exec := range tt.exec {
				_, err = e.Execute(ctx, utils.GenerateDBID(tt.database.Name, tt.database.Owner), exec.procedure, [][]any{exec.args})
				assert.NoError(t, err)
			}

		})
	}
}

type testExtension struct {
	requiredMetadata map[string]string
	methods          []string
}

func (t *testExtension) CreateInstance(ctx context.Context, metadata map[string]string) (engine.ExtensionInstance, error) {
	for k, v := range t.requiredMetadata {
		if metadata[k] != v {
			return nil, errors.New("metadata not found")
		}
	}

	return &testExtensionInstance{
		methods: t.methods,
	}, nil
}

type testExtensionInstance struct {
	methods []string
}

func (t *testExtensionInstance) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	for _, m := range t.methods {
		if m == method {
			return []any{1}, nil
		}
	}

	return nil, errors.New("method not found")
}

var testExtensions = map[string]engine.ExtensionInitializer{
	"erc20": &testExtension{
		requiredMetadata: map[string]string{
			"address": "0x1234",
		},
		methods: []string{
			"balanceOf",
		},
	},
}

func newMockRegister() *mockRegister {
	return &mockRegister{
		datasets: make(map[string]sql.Database),
	}
}

type mockRegister struct {
	datasets map[string]sql.Database
}

func (m *mockRegister) Register(ctx context.Context, name string, db sql.Database) error {
	_, ok := m.datasets[name]
	if ok {
		return errors.New("dataset already registered")
	}

	m.datasets[name] = db

	return nil
}

func (m *mockRegister) Unregister(ctx context.Context, name string) error {
	_, ok := m.datasets[name]
	if !ok {
		return errors.New("dataset not registered")
	}

	delete(m.datasets, name)

	return nil
}

var (
	procedureCreateUser = &types.Procedure{
		Name:   "create_user",
		Args:   []string{"$id", "$username", "$age"},
		Public: true,
		Statements: []string{
			"INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
		},
	}

	procedureCreatePost = &types.Procedure{
		Name:   "create_post",
		Args:   []string{"$id", "$title", "$content", "$date_string"},
		Public: true,
		Statements: []string{
			"INSERT INTO posts (id, title, content, author_id, post_date)VALUES ($id, $title, $content, (SELECT id FROM users WHERE address=@caller), $date_string);",
		},
	}
)

// deepEqual does a deep comparison, while considering empty slices as equal to nils.
func deepEqual(a, b any) bool {
	return cmp.Equal(a, b, cmpopts.EquateEmpty())
}
