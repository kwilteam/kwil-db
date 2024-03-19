package testdata

import (
	types "github.com/kwilteam/kwil-db/common"
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
		ProcedureRecursive,
		ProcedureRecursiveSneakyA,
		ProcedureRecursiveSneakyB,
	},
	Extensions: []*types.Extension{},
}
