package dataset_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/db/test"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/stretchr/testify/assert"
)

// TODO: test things that should not work, like calling a private procedure,
// deploying gibberish in a procedure, incorrect table names, etc.

func Test_Execute(t *testing.T) {
	type fields struct {
		ownerAddress            string
		callerAddress           string
		availableExtensions     []*testExt
		extensionInitialization []*types.Extension
		tables                  []*types.Table
		procedures              []*types.Procedure
	}

	defaultFields := fields{
		ownerAddress:            callerAddress,
		callerAddress:           callerAddress,
		availableExtensions:     testAvailableExtensions,
		extensionInitialization: testExtensions,
		tables:                  test_tables,
		procedures:              test_procedures,
	}
	_ = defaultFields

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
		// by default it is Execute(), but if we want to test Call() we can set this to true
		isCall         bool
		wantErr        bool
		wantBuilderErr bool
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
							"SELECT username, (SELECT count(*) FROM users) as num_users FROM users WHERE id = $id;",
							//"SELECT count(*) FROM users;",
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
		{
			name: "use extension that is not included in the extensions list",
			fields: fields{
				availableExtensions:     testAvailableExtensions,
				extensionInitialization: testExtensions,
				tables:                  test_tables,
				procedures: []*types.Procedure{
					{
						Name:   "use_ext",
						Args:   []string{"$name"},
						Public: true,
						Statements: []string{
							"$result = crypto.keccack256($name);",
						},
					},
				},
			},
			args: args{
				procedure: "use_ext",
				inputs: []map[string]interface{}{
					{
						"$name": "satoshi",
					},
				},
			},
			expectedOutputs: nil,
			wantErr:         true,
		},
		{
			name: "use extension that this server does not have an initializer for",
			fields: fields{
				availableExtensions: testAvailableExtensions,
				extensionInitialization: []*types.Extension{
					{
						Name:           "crypto",
						Initialization: map[string]string{},
						Alias:          "crypto",
					},
				},
				tables: test_tables,
				procedures: []*types.Procedure{
					{
						Name:   "use_ext",
						Args:   []string{"$name"},
						Public: true,
						Statements: []string{
							"$result = crypto.keccack256($name);",
						},
					},
				},
			},
			args: args{
				procedure: "use_ext",
				inputs: []map[string]interface{}{
					{
						"$name": "satoshi",
					},
				},
			},
			expectedOutputs: nil,
			wantErr:         true,
			wantBuilderErr:  true,
		},
		{
			name: "execute authenticated procedure without caller address should fail",
			fields: fields{
				availableExtensions:     testAvailableExtensions,
				extensionInitialization: testExtensions,
				tables:                  test_tables,
				procedures: []*types.Procedure{
					{
						Name:      "create_user",
						Args:      []string{},
						Public:    true,
						Modifiers: []types.Modifier{types.ModifierAuthenticated, types.ModifierView},
						Statements: []string{
							"SELECT * FROM users WHERE address = @caller;",
						},
					},
				},
			},
			args: args{
				procedure: "create_user",
				inputs:    []map[string]interface{}{},
			},
			expectedOutputs: nil,
			isCall:          true,
			wantErr:         true,
			wantBuilderErr:  false,
		},
		{
			name: "execute authenticated procedure with caller address should succeed",
			fields: fields{
				ownerAddress:            callerAddress,
				callerAddress:           callerAddress,
				availableExtensions:     testAvailableExtensions,
				extensionInitialization: testExtensions,
				tables:                  test_tables,
				procedures: []*types.Procedure{
					{
						Name:      "create_user",
						Args:      []string{},
						Public:    true,
						Modifiers: []types.Modifier{types.ModifierAuthenticated, types.ModifierView},
						Statements: []string{
							"SELECT * FROM users WHERE address = @caller;",
						},
					},
				},
			},
			args: args{
				procedure: "create_user",
				inputs:    []map[string]interface{}{},
			},
			expectedOutputs: []map[string]interface{}{},
			isCall:          true,
			wantErr:         false,
			wantBuilderErr:  false,
		},
		{
			name: "execute owner only procedure with non-owner caller address should fail",
			fields: fields{
				ownerAddress:            callerAddress,
				callerAddress:           "0xnotowner",
				availableExtensions:     testAvailableExtensions,
				extensionInitialization: testExtensions,
				tables:                  test_tables,
				procedures: []*types.Procedure{
					{
						Name:      "create_user",
						Args:      []string{},
						Public:    true,
						Modifiers: []types.Modifier{types.ModifierOwner},
						Statements: []string{
							"INSERT INTO users (id, username, age, address) VALUES (1, 'test_username', 20, @caller);",
						},
					},
				},
			},
			args: args{
				procedure: "create_user",
				inputs:    []map[string]interface{}{},
			},
			expectedOutputs: nil,
			isCall:          false,
			wantErr:         true,
			wantBuilderErr:  false,
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
				WithDatastore(database).
				Named(datasetName).OwnedBy(tt.fields.ownerAddress).
				Build(ctx)
			if tt.wantBuilderErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
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

			var outputs []map[string]interface{}
			if tt.isCall {
				if len(tt.args.inputs) == 0 {
					tt.args.inputs = []map[string]interface{}{nil}
				}

				outputs, err = ds.Call(ctx, tt.args.procedure, tt.args.inputs[0], &dataset.TxOpts{
					Caller: tt.fields.callerAddress,
				})
			} else {
				outputs, err = ds.Execute(ctx, tt.args.procedure, tt.args.inputs, &dataset.TxOpts{
					Caller: tt.fields.callerAddress,
				})
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Dataset.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.EqualValues(t, tt.expectedOutputs, outputs, fmt.Sprintf("expected %v, got %v", tt.expectedOutputs, outputs))
		})
	}
}
