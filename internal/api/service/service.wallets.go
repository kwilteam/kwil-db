package service

import (
	"context"

	//v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	types "github.com/kwilteam/kwil-db/pkg/types/db"
)

func (s *Service) GetWalletRole(ctx context.Context, req *v0.GetWalletRoleRequest) (*v0.GetWalletRoleResponse, error) {

	myRole := types.Role{
		Name: "admin",
		Permissions: types.Permissions{
			DDL:                  true,
			ParamaterizedQueries: []string{"test_insert"},
		},
	}

	return &v0.GetWalletRoleResponse{
		Name:        myRole.Name,
		Permissions: &v0.GetWalletRoleResponsePerms{Ddl: myRole.Permissions.DDL, Queries: myRole.Permissions.ParamaterizedQueries},
	}, nil
}
