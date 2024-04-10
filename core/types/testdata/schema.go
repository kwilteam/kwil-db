package testdata

import (
	"github.com/kwilteam/kwil-db/core/types"
)

// TestSchema is a test schema that mocks a social media application
var TestSchema = &types.Schema{
	Name:  "test_schema",
	Owner: []byte("test_owner"),
	Tables: []*types.Table{
		TableUsers,
		TablePosts,
	},
	Actions: []*types.Action{
		ActionCreateUser,
		ActionCreatePost,
		ActionGetUserByAddress,
		ActionGetPosts,
		ActionAdminDeleteUser,
		ActionCallsPrivate,
		ActionPrivate,
		ActionRecursive,
		ActionRecursiveSneakyA,
		ActionRecursiveSneakyB,
	},
	Procedures: []*types.Procedure{
		ProcCreateUser,
		ProcGetUserByAddress,
		ProcGetUsersByAge,
	},
	Extensions: []*types.Extension{},
}
