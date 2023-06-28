package engine_test

import (
	"context"
	"errors"
	"testing"

	engine "github.com/kwilteam/kwil-db/pkg/engine/newengine"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/log"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

func Test_CreateDataset(t *testing.T) {
	type dbStruct struct {
		dbName string
		owner  string
		schema *engine.Schema
	}
	tests := []struct {
		name     string
		database dbStruct
		wantErr  bool
	}{
		{
			name: "create a dataset",
			database: dbStruct{
				dbName: "test_db",
				owner:  "test_owner",
				schema: &engine.Schema{
					Extensions: []*types.Extension{
						{
							Name: "erc20",
							Initialization: map[string]string{
								"address": "0x1234",
							},
							Alias: "usdc",
						},
					},
					Tables: []*types.Table{
						&testdata.Table_users,
						&testdata.Table_posts,
					},
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
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			opener := newTestDBOpener()
			engine.DbOpener = opener
			defer opener.Teardown()

			e, err := engine.Open(ctx,
				engine.WithExtensions(testExtensions),
			)
			if err != nil {
				t.Fatal(err)
			}

			err = e.CreateDataset(ctx, tt.database.dbName, tt.database.owner, tt.database.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Engine.CreateDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// check that the dataset was created
			_, err = e.GetDataset(ctx, utils.GenerateDBID(tt.database.dbName, tt.database.owner))
			if err != nil {
				t.Fatal(err)
			}

		})
	}
}

func newTestDBOpener() *testDbOpener {
	return &testDbOpener{
		teardowns: make([]func() error, 0),
	}
}

// testDbOpener creates real sqlite databases that can be used for testing
type testDbOpener struct {
	teardowns []func() error
}

func (t *testDbOpener) Open(name, path string, l log.Logger) (engine.Datastore, error) {
	ds, teardown, err := sqlTesting.OpenTestDB(name)
	if err != nil {
		return nil, err
	}

	t.teardowns = append(t.teardowns, teardown)

	return &datastoreAdapter{ds}, nil
}

func (t *testDbOpener) Teardown() error {
	var errs []error
	for _, teardown := range t.teardowns {
		err := teardown()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type datastoreAdapter struct {
	sqlTesting.TestSqliteClient
}

func (d *datastoreAdapter) Prepare(query string) (engine.Statement, error) {
	return d.TestSqliteClient.Prepare(query)
}

func (d *datastoreAdapter) Savepoint() (engine.Savepoint, error) {
	return d.TestSqliteClient.Savepoint()
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
