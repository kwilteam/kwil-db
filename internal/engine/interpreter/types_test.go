package interpreter

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

func init() {
	// we parse all bodies to create the AST
	for _, action := range test_AllActions {
		res, err := parse.ParseActionBodyWithoutValidation(action.RawBody)
		if res.Errs.Err() != nil {
			panic(res.Errs.Err())
		}
		if err != nil {
			panic(err)
		}

		action.Body = res.AST
	}
}

// actions
var (
	test_AllActions = []*Action{
		test_ACTION_CreateUser,
		test_ACTION_ListUsers,
		test_ACTION_GetUserByName,
	}
	test_ACTION_CreateUser = &Action{
		Name: "create_user",
		Parameters: []*NamedType{
			{
				Name: "$name",
				Type: types.TextType.Copy(),
			},
			{
				Name: "$age",
				Type: types.IntType.Copy(),
			},
		},
		Public: true,
		RawBody: `
		INSERT INTO users (id, name, age)
		VALUES (
			uuid_generate_v5('c7b6a54c-392c-48f9-803d-31cb97e76052'::uuid, @txid),
			$name,
			$age
		);
		`,
	}
	test_ACTION_ListUsers = &Action{
		Name:       "list_users",
		Parameters: []*NamedType{},
		Public:     true,
		RawBody: `
		RETURN SELECT id, name, age
		FROM users;
		`,
		Modifiers: []Modifier{
			ModifierView,
			ModifierOwner,
		},
		Returns: &ActionReturn{
			IsTable: true,
			Fields: []*NamedType{
				{
					Name: "id",
					Type: types.UUIDType.Copy(),
				},
				{
					Name: "name",
					Type: types.TextType.Copy(),
				},
				{
					Name: "age",
					Type: types.IntType.Copy(),
				},
			},
		},
	}

	test_ACTION_GetUserByName = &Action{
		Name: "get_user_by_name",
		Parameters: []*NamedType{
			{
				Name: "$name",
				Type: types.TextType.Copy(),
			},
		},
		Public: true,
		RawBody: `
		FOR $row IN SELECT id, age
		FROM users
		WHERE name = $name {
			RETURN $row.id, $row.age;
		}
		`,
		Modifiers: []Modifier{
			ModifierView,
		},
		Returns: &ActionReturn{
			IsTable: false,
			Fields: []*NamedType{
				{
					Name: "id",
					Type: types.UUIDType.Copy(),
				},
				{
					Name: "age",
					Type: types.IntType.Copy(),
				},
			},
		},
	}
)

// tables

var (
	test_AllTables = []string{
		test_TABLE_Users,
	}

	test_TABLE_Users = `
	CREATE TABLE users (
		id UUID PRIMARY KEY,
		name TEXT NOT NULL CHECK (name <> '' AND length(name) <= 100),
 		age INT CHECK (age >= 0)
	);
	`
)
