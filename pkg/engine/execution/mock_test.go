package execution_test

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/execution"
)

// this file contains mocks

type mockDatastore struct {
	Identifier         string
	createsNonMutative bool // default false
}

func (m *mockDatastore) Prepare(ctx context.Context, uery string) (execution.PreparedStatement, error) {

	return &mockPreparedStatement{
		mutative: !m.createsNonMutative,
	}, nil
}

func (m *mockDatastore) Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error) {
	return []map[string]any{}, nil
}

type mockPreparedStatement struct {
	mutative bool
}

func (m *mockPreparedStatement) Close() error {
	return nil
}

func (m *mockPreparedStatement) Execute(ctx context.Context, args map[string]any) ([]map[string]any, error) {
	return nil, nil
}

func (m *mockPreparedStatement) IsMutative() bool {
	return m.mutative
}

var (
	testExtensions = map[string]execution.Initializer{
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

	testProcedures = map[string]*execution.Procedure{
		"publicProcedure": {
			Name: "publicProcedure",
			Parameters: []string{
				"$arg1",
				"$arg2",
			},
			Scoping: execution.ProcedureScopingPublic,
			Body: []*execution.InstructionExecution{
				{
					Instruction: execution.OpExtensionExecute,
					Args: []any{
						"usdc",
						"balanceOf",
						[]string{"$arg1"},
						[]string{"$res1"},
					},
				},
				{
					Instruction: execution.OpDMLExecute,
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
			Scoping: execution.ProcedureScopingPrivate,
			Body: []*execution.InstructionExecution{
				{
					Instruction: execution.OpDMLExecute,
					Args: []any{
						"has_balance",
					},
				},
			},
		},
	}

	testExecutionOpts = []execution.ExecutionOpt{
		execution.WithCaller(&mockUser{}),
		execution.WithDatasetID("xDBID"),
	}

	testLoadCommand = []*execution.InstructionExecution{
		{
			Instruction: execution.OpSetVariable,
			Args: []any{
				"$usdc_address",
				"0x12345678901",
			},
		},
		{
			Instruction: execution.OpExtensionInitialize,
			Args: []any{
				"erc20",
				"usdc",
				map[string]string{
					"address": "$usdc_address",
				},
			},
		},
		{
			Instruction: execution.OpDMLPrepare,
			Args: []any{
				"update_balance",
				"UPDATE balances SET balance = $res1 WHERE address = $arg1",
			},
		},
		{
			Instruction: execution.OpDMLPrepare,
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

func (m *mockInitializer) Initialize(ctx context.Context, metadata map[string]string) (execution.InitializedExtension, error) {
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

type mockUser struct{}

func (m *mockUser) Address() string {
	return "0xCaller"
}

func (m *mockUser) Bytes() []byte {
	return []byte("000xPUBKEY")
}

func (m *mockUser) PubKey() []byte {
	return []byte("0xPUBKEY")
}

type mockEvaluater struct {
	val any
}

func newMockEvaluater(returnVal any) *mockEvaluater {
	return &mockEvaluater{
		val: returnVal,
	}
}

func (m *mockEvaluater) Evaluate(expr string, values map[string]any) (any, error) {
	return m.val, nil
}

func (m *mockEvaluater) Close() error {
	return nil
}
