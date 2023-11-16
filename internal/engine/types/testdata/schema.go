package testdata

import (
	"github.com/kwilteam/kwil-db/internal/engine/types"
)

// TestSchema is a test schema that mocks a social media application
var TestSchema = &types.Schema{
	Name:  "test_schema",
	Owner: []byte("test_owner"),
	Tables: []*types.Table{
		TableUsers,
		TablePosts,
	},
	Procedures: []*types.Procedure{
		ProcedureCreateUser,
		ProcedureCreatePost,
		ProcedureGetUserByAddress,
		ProcedureGetPosts,
		ProcedureAdminDeleteUser,
		ProcedureCallsPrivate,
		ProcedurePrivate,
	},
	Extensions: []*types.Extension{},
}
