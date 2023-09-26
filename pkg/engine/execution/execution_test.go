package execution_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/execution"
)

func TestEngine_ExecuteProcedure(t *testing.T) {
	type fields struct {
		availableExtensions map[string]execution.Initializer
		procedures          map[string]*execution.Procedure
		loadCommand         []*execution.InstructionExecution
		db                  execution.Datastore
		evaluater           execution.Evaluater
	}
	type args struct {
		ctx  context.Context
		name string
		args []any
		opts []execution.ExecutionOpt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []map[string]any
		wantErr error // if nil, no error is expected
	}{
		// Extension tests
		{
			name: "executing an extension without initializing it should fail",
			fields: fields{
				availableExtensions: testExtensions,
				procedures:          testProcedures,
				db:                  &mockDatastore{},
			},
			args: args{
				ctx:  context.Background(),
				name: "publicProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: execution.ErrUnitializedExtension,
		},
		{
			name: "executing an extension with initialization in the procedure should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"erc20_procedure": {
						Name:       "erc20_procedure",
						Parameters: []string{"$address", "$arg2"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpExtensionInitialize,
								Args:        []any{"erc20", "usdc", map[string]string{"address": "$address"}},
							},
							{
								Instruction: execution.OpExtensionExecute,
								Args:        []any{"usdc", "balanceOf", []string{"$address"}, []string{"$wallet_balance"}},
							},
						},
					},
				},
				db: &mockDatastore{},
			},
			args: args{
				ctx:  context.Background(),
				name: "erc20_procedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "executing an extension with initialization in the procedure should fail if the extension is not available",
			fields: fields{
				availableExtensions: map[string]execution.Initializer{},
				procedures: map[string]*execution.Procedure{
					"erc20_procedure": {
						Name:       "erc20_procedure",
						Parameters: []string{"$address", "$arg2"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpExtensionInitialize,
								Args:        []any{"erc20", "usdc", map[string]string{"address": "$address"}},
							},
							{
								Instruction: execution.OpExtensionExecute,
								Args:        []any{"usdc", "balanceOf", []string{"$address"}, []string{"$wallet_balance"}},
							},
						},
					},
				},
				db: &mockDatastore{},
			},
			args: args{
				ctx:  context.Background(),
				name: "erc20_procedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: execution.ErrUnknownExtension,
		},
		{
			name: "executing an extension with initialization in the load command should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"erc20_procedure": {
						Name:       "erc20_procedure",
						Parameters: []string{"$address"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpExtensionExecute,
								Args:        []any{"usdc", "balanceOf", []string{"$address"}, []string{"$wallet_balance"}},
							},
						},
					},
				},
				db: &mockDatastore{},
				loadCommand: []*execution.InstructionExecution{
					{
						Instruction: execution.OpSetVariable,
						Args:        []any{"!address", "0x123"},
					},
					{
						Instruction: execution.OpExtensionInitialize,
						Args:        []any{"erc20", "usdc", map[string]string{"address": "!address"}},
					},
				},
			},
			args: args{
				ctx:  context.Background(),
				name: "erc20_procedure",
				args: []any{"0x123"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: nil,
		},
		// DML tests
		{
			name: "executing a statement without preparing it should fail",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpDMLExecute,
								Args:        []any{"dml_statement"},
							},
						},
					},
				},
				db: &mockDatastore{},
			},
			args: args{
				ctx:  context.Background(),
				name: "dmlProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: execution.ErrUnknownPreparedStatement,
		},
		{
			name: "executing a statement after preparing it should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpDMLPrepare,
								Args:        []any{"dml_statement", "SELECT * FROM users WHERE id = $arg1"},
							},
							{
								Instruction: execution.OpDMLExecute,
								Args:        []any{"dml_statement"},
							},
						},
					},
				},
				db: &mockDatastore{},
			},
			args: args{
				ctx:  context.Background(),
				name: "dmlProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "executing a statement after preparing in the load command should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpDMLExecute,
								Args:        []any{"dml_statement"},
							},
						},
					},
				},
				db: &mockDatastore{},
				loadCommand: []*execution.InstructionExecution{
					{
						Instruction: execution.OpDMLPrepare,
						Args:        []any{"dml_statement", "SELECT * FROM users WHERE id = $arg1"},
					},
				},
			},
			args: args{
				ctx:  context.Background(),
				name: "dmlProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "executing a mutative statement, setting context to immutable should fail",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpDMLExecute,
								Args:        []any{"dml_statement"},
							},
						},
					},
				},
				db: &mockDatastore{},
				loadCommand: []*execution.InstructionExecution{
					{
						Instruction: execution.OpDMLPrepare,
						Args:        []any{"dml_statement", "INSERT INTO users (id, username, age) VALUES (1, 'test', 20)"},
					},
				},
			},
			args: args{
				ctx:  context.Background(),
				name: "dmlProcedure",
				args: []any{},
				opts: append(testExecutionOpts, execution.NonMutative()),
			},
			want:    nil,
			wantErr: execution.ErrMutativeStatement,
		},
		{
			name: "executing a non-mutative statement, setting context to immutable should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpDMLExecute,
								Args:        []any{"dml_statement"},
							},
						},
					},
				},
				db: &mockDatastore{
					createsNonMutative: true,
				},
				loadCommand: []*execution.InstructionExecution{
					{
						Instruction: execution.OpDMLPrepare,
						Args:        []any{"dml_statement", "INSERT INTO users (id, username, age) VALUES (1, 'test', 20)"},
					},
				},
			},
			args: args{
				ctx:  context.Background(),
				name: "dmlProcedure",
				args: []any{},
				opts: append(testExecutionOpts, execution.NonMutative()),
			},
			want:    nil,
			wantErr: nil,
		},
		// Procedure tests
		{
			name: "executing a procedure with the wrong number of arguments should fail",
			fields: fields{
				availableExtensions: testExtensions,
				procedures:          testProcedures,
				db:                  &mockDatastore{},
				loadCommand:         testLoadCommand,
			},
			args: args{
				ctx:  context.Background(),
				name: "publicProcedure",
				args: []any{"0x123", "0xabc", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: execution.ErrIncorrectNumArgs,
		},
		{
			name: "executing a procedure with the correct number of arguments should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures:          testProcedures,
				db:                  &mockDatastore{},
				loadCommand:         testLoadCommand,
			},
			args: args{
				ctx:  context.Background(),
				name: "publicProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "executing a private procedure directly should fail",
			fields: fields{
				availableExtensions: testExtensions,
				procedures:          testProcedures,
				db:                  &mockDatastore{},
				loadCommand:         testLoadCommand,
			},
			args: args{
				ctx:  context.Background(),
				name: "privateProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: execution.ErrScopingViolation,
		},
		{
			name: "executing a private procedure indirectly should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"publicProcedure": {
						Name:       "publicProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpProcedureExecute,
								Args:        []any{"privateProcedure", []string{"$arg1", "$arg2"}},
							},
						},
					},
					"privateProcedure": {
						Name:       "privateProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    execution.ProcedureScopingPrivate,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpSetVariable,
								Args:        []any{"!var1", "hello"},
							},
						},
					},
				},
				db:          &mockDatastore{},
				loadCommand: testLoadCommand,
			},
			args: args{
				ctx:  context.Background(),
				name: "publicProcedure",
				args: []any{"0x123", "0xabc"},
				opts: testExecutionOpts,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "basic evaluatable procedure should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*execution.Procedure{
					"evaluatableProcedure": {
						Name:       "evaluatableProcedure",
						Parameters: []string{"$arg1"},
						Scoping:    execution.ProcedureScopingPublic,
						Body: []*execution.InstructionExecution{
							{
								Instruction: execution.OpEvaluatable,
								Args:        []any{"SELECT $arg1", "$arg1"},
							},
						},
					},
				},
				db:          &mockDatastore{},
				loadCommand: testLoadCommand,
				evaluater:   newMockEvaluater("result"),
			},
			args: args{
				ctx:  context.Background(),
				name: "evaluatableProcedure",
				args: []any{"0x123"},
				opts: testExecutionOpts,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.evaluater == nil {
				tt.fields.evaluater = newMockEvaluater("default_mock_evaluater")
			}

			e, err := execution.NewEngine(context.Background(), tt.fields.db, tt.fields.evaluater, &execution.EngineOpts{
				Extensions: tt.fields.availableExtensions,
				Procedures: tt.fields.procedures,
				LoadCmd:    tt.fields.loadCommand,
			})
			if err != nil {
				t.Errorf("Engine.ExecuteProcedure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer e.Close()

			got, err := e.ExecuteProcedure(tt.args.ctx, tt.args.name, tt.args.args, tt.args.opts...)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Engine.ExecuteProcedure() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				return
			}

			if err != nil {
				t.Errorf("Engine.ExecuteProcedure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Engine.ExecuteProcedure() = %v, want %v", got, tt.want)
			}
		})
	}
}
