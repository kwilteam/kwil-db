package dataset_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/db/test"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/stretchr/testify/assert"
)

// TODO: test things that should not work, like calling a private procedure,
// deploying gibberish in a procedure, incorrect table names, etc.

func Test_Execute(t *testing.T) {
	type fields struct {
		availableExtensions     []*testExt
		extensionInitialization []*types.Extension
		tables                  []*types.Table
		procedures              []*types.Procedure
	}

	defaultFields := fields{
		availableExtensions:     testAvailableExtensions,
		extensionInitialization: testExtensions,
		tables:                  test_tables,
		procedures:              test_procedures,
	}

	type args struct {
		procedure string
		inputs    []map[string]interface{}
		finisher  func(*dataset.Dataset) error
	}

	tests := []struct {
		name            string
		fields          fields
		args            args
		expectedOutputs []map[string]interface{}
		wantErr         bool
	}{
		{
			name:   "execute a dml procedure successfully",
			fields: defaultFields,
			args: args{
				procedure: "create_user",
				inputs: []map[string]interface{}{
					{
						"$id":       "1",
						"$username": "test_username",
						"$age":      20,
					},
				},
			},
			expectedOutputs: []map[string]interface{}{},
			wantErr:         false,
		},
		{
			name:   "execute a procedure with an extension successfully",
			fields: defaultFields,
			args: args{
				procedure: "get_time",
				inputs:    []map[string]interface{}{},
			},
			expectedOutputs: nil,
			wantErr:         false,
		},
		{
			name:   "violate foreign key constraint",
			fields: defaultFields,
			args: args{
				procedure: "create_post",
				inputs: []map[string]interface{}{
					{
						"$id":        "1",
						"$title":     "test_title",
						"$content":   "test_content",
						"$author_id": "20485",
					},
				},
			},
			expectedOutputs: nil,
			wantErr:         true,
		},
		{
			name:   "execute nested procedure that returns data successfully",
			fields: defaultFields,
			args: args{
				procedure: "create_post_and_user",
				inputs: []map[string]interface{}{
					{
						"$id":        "1",
						"$title":     "test_title",
						"$content":   "test_content",
						"$author_id": "1",
						"$username":  "test_username",
						"$age":       20,
					},
				},
			},
			expectedOutputs: []map[string]interface{}{
				{
					"username":       "test_username",
					"wallet_address": callerAddress,
					"title":          "test_title",
					"content":        "test_content",
				},
			},
			wantErr: false,
		},
		{
			name: "batch execute and return the final output",
			fields: fields{
				availableExtensions:     testAvailableExtensions,
				extensionInitialization: testExtensions,
				tables:                  test_tables,
				procedures: []*types.Procedure{
					{
						Name:   "create_user_manual",
						Args:   []string{"$id", "$username", "$age", "$address"},
						Public: true,
						Statements: []string{
							"INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, $address);",
							`SELECT username, (SELECT count(*) FROM users) as num_users FROM users WHERE id = $id;`,
						},
					},
				},
			},
			args: args{
				procedure: "create_user_manual",
				inputs: []map[string]interface{}{
					{
						"$id":       "1",
						"$username": "test_username",
						"$age":      20,
						"$address":  "0x123",
					},
					{
						"$id":       "2",
						"$username": "test_username2",
						"$age":      20,
						"$address":  "0x456",
					},
				},
			},
			expectedOutputs: []map[string]interface{}{
				{
					"username":  "test_username2",
					"num_users": int64(2), // we get num users to make sure the first insert was successful
				},
			},
			wantErr: false,
		},
		{
			name: "failed batch insert will revert all inserts",
			fields: fields{
				availableExtensions:     testAvailableExtensions,
				extensionInitialization: testExtensions,
				tables:                  test_tables,
				procedures: []*types.Procedure{
					{
						Name:   "create_user_manual",
						Args:   []string{"$id", "$username", "$age", "$address"},
						Public: true,
						Statements: []string{
							"INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, $address);",
						},
					},
				},
			},
			args: args{
				procedure: "create_user_manual",
				inputs: []map[string]interface{}{
					{
						"$id":       "1",
						"$username": "test_username",
						"$age":      20,
						"$address":  "0x123",
					},
					{
						"$id":       "2abc", // this will fail
						"$username": "test_username2",
						"$age":      20,
						"$address":  "0x456",
					},
				},
				finisher: func(database *dataset.Dataset) error {
					results, err := database.Query(context.Background(), "SELECT * FROM users;", nil)
					if err != nil {
						return err
					}

					if len(results) != 0 {
						return fmt.Errorf("expected no results, got %d", len(results))
					}

					return nil
				},
			},
			expectedOutputs: nil,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			database, teardown, err := test.OpenTestDB(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer teardown()

			availableExtensions := map[string]dataset.Initializer{}
			for _, etx := range tt.fields.availableExtensions {
				availableExtensions[etx.name] = etx
			}

			ds, err := dataset.Builder().
				WithTables(tt.fields.tables...).
				WithProcedures(tt.fields.procedures...).
				WithInitializers(availableExtensions).
				WithExtensions(tt.fields.extensionInitialization...).
				WithDatastore(databaseWrapper{database}).
				Named(datasetName).OwnedBy(callerAddress).
				Build(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer ds.Delete()

			defer func() {
				if tt.args.finisher != nil {
					err = tt.args.finisher(ds)
					if err != nil {
						t.Fatal(err)
					}
				}
			}()

			outputs, err := ds.Execute(ctx, tt.args.procedure, tt.args.inputs, &dataset.TxOpts{
				Caller: callerAddress,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Dataset.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.EqualValues(t, tt.expectedOutputs, outputs, fmt.Sprintf("expected %v, got %v", tt.expectedOutputs, outputs))
		})
	}
}

type databaseWrapper struct {
	*db.DB
}

func (d databaseWrapper) Prepare(stmt string) (dataset.Statement, error) {
	return d.DB.Prepare(stmt)
}

func (d databaseWrapper) Savepoint() (dataset.Savepoint, error) {
	return d.DB.Savepoint()
}
