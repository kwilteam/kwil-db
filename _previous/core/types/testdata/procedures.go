package testdata

import "github.com/kwilteam/kwil-db/core/types"

var (
	ProcCreateUser = &types.Procedure{
		Name: "proc_create_user",
		Parameters: []*types.ProcedureParameter{
			{Name: "$id", Type: types.IntType},
			{Name: "$username", Type: types.TextType},
			{Name: "$age", Type: types.IntType},
		},
		Public: true,
		Body:   "INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
	}

	ProcGetUserByAddress = &types.Procedure{
		Name: "proc_get_user_by_address",
		Parameters: []*types.ProcedureParameter{
			{Name: "$address", Type: types.TextType},
		},
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Body: `
		for $row in SELECT id, username, age FROM users WHERE address = $address {
			return $row.id, $row.username, $row.age;
		}
		error('User not found');
		`,
		Returns: &types.ProcedureReturn{
			Fields: []*types.NamedType{
				{
					Name: "id",
					Type: types.IntType,
				},
				{
					Name: "username",
					Type: types.TextType,
				},
				{
					Name: "age",
					Type: types.IntType,
				},
			},
		},
	}

	ProcGetUsersByAge = &types.Procedure{
		Name: "proc_get_users_by_age",
		Parameters: []*types.ProcedureParameter{
			{Name: "$age", Type: types.IntType},
		},
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Body: `
		return SELECT id, username, age FROM users WHERE age = $age;
		`,
		Returns: &types.ProcedureReturn{
			IsTable: true,
			Fields: []*types.NamedType{
				{Name: "id", Type: types.IntType},
				{Name: "username", Type: types.TextType},
				{Name: "age", Type: types.IntType},
			},
		},
	}
)
