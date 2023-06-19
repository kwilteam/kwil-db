package dataset2

import "github.com/kwilteam/kwil-db/pkg/engine/dto"

// TODO: this is a TEST schema to show to Gavin.

type Schema struct {
	Tables     []*dto.Table
	OnLoad     []*OpCodeExecution
	OnDeploy   []*OpCodeExecution
	Procedures []*Procedure
}

var (
	exampleSchema = &Schema{
		Tables: []*dto.Table{},
		OnLoad: []*OpCodeExecution{
			{
				OpCode: OpCodeDMLPrepare,
				Args: []any{
					"insert_user",
					"INSERT INTO users (id, name) VALUES ($id, $name)",
				},
			},
			{
				OpCode: OpCodeSetVariable,
				Args: []any{
					"!usdc_address",
					"0x1234",
				},
			},
			{
				OpCode: OpCodeExtensionInitialize,
				Args: []any{
					"erc20",
					"usdc",
					map[string]string{
						"address": "!usdc_address",
					},
				},
			},
		},
		/*
			constructor($constructor_name) {
				INSERT INTO users (id, name) VALUES (1, 'gavin')
				INSERT INTO users (id, name) VALUES (2, $constructor_name)
			}
		*/
		OnDeploy: []*OpCodeExecution{
			{
				OpCode: OpCodeDMLPrepare,
				Args: []any{
					"!insert_user_1",
					"INSERT INTO users (id, name) VALUES (1, 'Gavin')",
				},
			},
			{
				OpCode: OpCodeDMLPrepare,
				Args: []any{
					"!insert_user_2",
					"INSERT INTO users (id, name) VALUES (2, $constructor_name)", // $constructor_name is a variable that must be passed into a constructor
				},
			},
			{
				OpCode: OpCodeDMLExecute,
				Args: []any{
					"!insert_user_1",
				},
			},
			{
				OpCode: OpCodeDMLExecute,
				Args: []any{
					"!insert_user_2",
				},
			},
		},
		Procedures: []*Procedure{
			{
				Name: "create_user",
				Parameters: []string{
					"$id",
					"$name",
				},
				Scoping: ProcedureScopingPublic,
				Body: []*OpCodeExecution{
					{
						OpCode: OpCodeDMLExecute,
						Args: []any{
							"insert_user",
						},
					},
				},
			},
		},
	}
)

var (
	schema2 = &Schema{
		Tables: []*dto.Table{},
		OnLoad: []*OpCodeExecution{
			{
				OpCode: OpCodeSetVariable,
				Args: []any{
					"!0xzabxjaska",
					"0x1234",
				},
			},
			{
				OpCode: OpCodeExtensionInitialize,
				Args: []any{
					"erc20",
					"usdc",
					map[string]string{
						"address": "!0xzabxjaska",
					},
				},
			},
		},
		Procedures: []*Procedure{
			{
				Name: "create_user",
				Parameters: []string{
					"$id",
					"$name",
				},
				Scoping: ProcedureScopingPublic,
				Body: []*OpCodeExecution{
					{
						OpCode: OpCodeSetVariable,
						Args: []any{
							"!wallet_address",
							"0xabc",
						},
					},
					{
						OpCode: OpCodeExtensionExecute,
						Args: []any{
							"usdc",
							"balance",
							[]string{
								"!wallet_address",
							},
							[]string{
								"$balance",
							},
						},
					},
					{
						OpCode: OpCodeSetVariable,
						Args: []any{
							"!usdc_address",
							"0x5678",
						},
					},
					{
						OpCode: OpCodeExtensionInitialize,
						Args: []any{
							"erc20",
							"usdc",
							map[string]string{
								"address": "!usdc_address",
							},
						},
					},
				},
			},
		},
	}
)
