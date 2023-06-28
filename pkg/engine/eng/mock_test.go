package eng_test

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/eng"
)

// this file contains mocks

type mockDatastore struct {
	Identifier string
}

func (m *mockDatastore) Prepare(query string) (eng.PreparedStatement, error) {
	return &mockPreparedStatement{}, nil
}

type mockPreparedStatement struct {
}

func (m *mockPreparedStatement) Close() error {
	return nil
}

func (m *mockPreparedStatement) Execute(ctx context.Context, args map[string]any) ([]map[string]any, error) {
	return nil, nil
}

var (
	testExtensions = map[string]eng.Initializer{
		"erc20": &mockInitializer{
			methodReturns: map[string][]any{
				"balanceOf": {uint64(100)},
			},
		},
		"erc721": &mockInitializer{
			methodReturns: map[string][]any{
				"owner": {"0x123JPEG_HOLDER"},
			},
		},
	}

	testProcedures = map[string]*eng.Procedure{
		"publicProcedure": {
			Name: "publicProcedure",
			Parameters: []string{
				"$arg1",
				"$arg2",
			},
			Scoping: eng.ProcedureScopingPublic,
			Body: []*eng.InstructionExecution{
				{
					Instruction: eng.OpExtensionExecute,
					Args: []any{
						"usdc",
						"balanceOf",
						[]string{"$arg1"},
						[]string{"$res1"},
					},
				},
				{
					Instruction: eng.OpDMLExecute,
					Args: []any{
						"update_balance",
					},
				},
			},
		},
		"privateProcedure": {
			Name: "privateProcedure",
			Parameters: []string{
				"$arg1",
				"$arg2",
			},
			Scoping: eng.ProcedureScopingPrivate,
			Body: []*eng.InstructionExecution{
				{
					Instruction: eng.OpDMLExecute,
					Args: []any{
						"has_balance",
					},
				},
			},
		},
	}

	testExecutionOpts = []eng.ExecutionOpt{
		eng.WithCaller("0xCaller"),
		eng.WithDatasetID("xDBID"),
	}

	testLoadCommand = []*eng.InstructionExecution{
		{
			Instruction: eng.OpSetVariable,
			Args: []any{
				"$usdc_address",
				"0x12345678901",
			},
		},
		{
			Instruction: eng.OpExtensionInitialize,
			Args: []any{
				"erc20",
				"usdc",
				map[string]string{
					"address": "$usdc_address",
				},
			},
		},
		{
			Instruction: eng.OpDMLPrepare,
			Args: []any{
				"update_balance",
				"UPDATE balances SET balance = $res1 WHERE address = $arg1",
			},
		},
		{
			Instruction: eng.OpDMLPrepare,
			Args: []any{
				"has_balance",
				"SELECT balance FROM balances WHERE address = $arg1",
			},
		},
	}
)

type mockInitializer struct {
	methodReturns map[string][]any
}

func (m *mockInitializer) Initialize(ctx context.Context, metadata map[string]string) (eng.InitializedExtension, error) {
	return &mockInitializedExtension{
		methodReturns: m.methodReturns,
	}, nil
}

type mockInitializedExtension struct {
	methodReturns map[string][]any
}

func (m *mockInitializedExtension) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	return m.methodReturns[method], nil
}
