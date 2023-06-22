package eng_test

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/eng"
)

func TestEngine_ExecuteProcedure(t *testing.T) {
	type fields struct {
		availableExtensions map[string]eng.Initializer
		procedures          map[string]*eng.Procedure
		loadCommand         []*eng.InstructionExecution
		db                  eng.Datastore
	}
	type args struct {
		ctx  context.Context
		name string
		args []any
		opts []eng.ExecutionOpt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.Reader
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
			wantErr: eng.ErrUnitializedExtension,
		},
		{
			name: "executing an extension with initialization in the procedure should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*eng.Procedure{
					"erc20_procedure": {
						Name:       "erc20_procedure",
						Parameters: []string{"$address", "$arg2"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpExtensionInitialize,
								Args:        []any{"erc20", "usdc", map[string]string{"address": "$address"}},
							},
							{
								Instruction: eng.OpExtensionExecute,
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
				availableExtensions: map[string]eng.Initializer{},
				procedures: map[string]*eng.Procedure{
					"erc20_procedure": {
						Name:       "erc20_procedure",
						Parameters: []string{"$address", "$arg2"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpExtensionInitialize,
								Args:        []any{"erc20", "usdc", map[string]string{"address": "$address"}},
							},
							{
								Instruction: eng.OpExtensionExecute,
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
			wantErr: eng.ErrUnknownExtension,
		},
		{
			name: "executing an extension with initialization in the load command should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*eng.Procedure{
					"erc20_procedure": {
						Name:       "erc20_procedure",
						Parameters: []string{"$address"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpExtensionExecute,
								Args:        []any{"usdc", "balanceOf", []string{"$address"}, []string{"$wallet_balance"}},
							},
						},
					},
				},
				db: &mockDatastore{},
				loadCommand: []*eng.InstructionExecution{
					{
						Instruction: eng.OpSetVariable,
						Args:        []any{"!address", "0x123"},
					},
					{
						Instruction: eng.OpExtensionInitialize,
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
				procedures: map[string]*eng.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpDMLExecute,
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
			wantErr: eng.ErrUnknownPreparedStatement,
		},
		{
			name: "executing a statement after preparing it should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*eng.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpDMLPrepare,
								Args:        []any{"dml_statement", "SELECT * FROM users WHERE id = $arg1"},
							},
							{
								Instruction: eng.OpDMLExecute,
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
				procedures: map[string]*eng.Procedure{
					"dmlProcedure": {
						Name:       "dmlProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpDMLExecute,
								Args:        []any{"dml_statement"},
							},
						},
					},
				},
				db: &mockDatastore{},
				loadCommand: []*eng.InstructionExecution{
					{
						Instruction: eng.OpDMLPrepare,
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
			wantErr: eng.ErrIncorrectNumArgs,
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
			wantErr: eng.ErrScopingViolation,
		},
		{
			name: "executing a private procedure indirectly should succeed",
			fields: fields{
				availableExtensions: testExtensions,
				procedures: map[string]*eng.Procedure{
					"publicProcedure": {
						Name:       "publicProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    eng.ProcedureScopingPublic,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpProcedureExecute,
								Args:        []any{"privateProcedure", []string{"$arg1", "$arg2"}},
							},
						},
					},
					"privateProcedure": {
						Name:       "privateProcedure",
						Parameters: []string{"$arg1", "$arg2"},
						Scoping:    eng.ProcedureScopingPrivate,
						Body: []*eng.InstructionExecution{
							{
								Instruction: eng.OpSetVariable,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := eng.NewEngine(context.Background(), tt.fields.db, &eng.EngineOpts{
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
