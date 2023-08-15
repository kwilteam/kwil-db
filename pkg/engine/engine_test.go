package engine_test

import (
	"context"
	"errors"
	"testing"

	engine "github.com/kwilteam/kwil-db/pkg/engine"
	engineTesting "github.com/kwilteam/kwil-db/pkg/engine/testing"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/stretchr/testify/assert"
)

var (
	testTables = []*types.Table{
		&testdata.Table_users,
		&testdata.Table_posts,
	}

	testProcedures = []*types.Procedure{
		&testdata.Procedure_create_user,
		&testdata.Procedure_create_post,
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

	e, teardown, err := engineTesting.NewTestEngine(ctx,
		engine.WithExtensions(testExtensions),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	_, err = e.CreateDataset(ctx, &types.Schema{
		Name:       "testdb1",
		Owner:      "0xSatoshi",
		Extensions: testInitializedExtensions,
		Tables:     testTables,
		Procedures: testProcedures,
	})
	if err != nil {
		t.Fatal(err)
	}

	// close the engine
	err = e.Close()
	if err != nil {
		t.Fatal(err)
	}

	e2, teardown2, err := engineTesting.NewTestEngine(ctx,
		engine.WithExtensions(testExtensions),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown2()

	// check if the dataset was created
	dataset, err := e2.GetDataset(ctx, utils.GenerateDBID("testdb1", "0xSatoshi"))
	if err != nil {
		t.Fatal(err)
	}

	// check if the dataset has the correct tables
	tables, err := dataset.ListTables(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.ElementsMatch(t, testTables, tables)

	// check if the dataset has the correct procedures
	procs := dataset.ListProcedures()
	assert.ElementsMatch(t, testProcedures, procs)

	// list the datasets
	datasets, err := e2.ListDatasets(ctx, "0xSatoshi")
	if err != nil {
		t.Fatal(err)
	}

	assert.ElementsMatch(t, []string{"testdb1"}, datasets)
}

func Test_CreateDataset(t *testing.T) {
	tests := []struct {
		name     string
		database *types.Schema
		wantErr  bool
	}{
		{
			name: "create a dataset with a variety of statements",
			database: &types.Schema{
				Name:       "test_db",
				Owner:      "test_owner",
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
				Owner:      "test_owner",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			e, teardown, err := engineTesting.NewTestEngine(ctx,
				engine.WithExtensions(testExtensions),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer teardown()

			_, err = e.CreateDataset(ctx, tt.database)
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
			if hasErr {
				assert.Error(t, err)
			} else {
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
			return []any{}, nil
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
