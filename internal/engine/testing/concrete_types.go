package testing

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/parse"
)

func init() {
	// we parse all bodies to create the AST
	for _, action := range []*engine.Action{
		ACTION_CreateUser,
	} {
		parse.ParseActionBodyWithoutValidation()
	}

}

var (
	ACTION_CreateUser = &engine.Action{
		Name: "create_user",
		Parameters: []*engine.NamedType{
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
)
