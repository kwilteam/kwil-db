package apisvc

import (
	"context"

	types "kwil/pkg/types/db"
	apipb "kwil/x/proto/apisvc"
)

func (s *Service) GetWalletRole(ctx context.Context, req *apipb.GetWalletRoleRequest) (*apipb.GetWalletRoleResponse, error) {

	myRole := types.Role{
		Name: "admin",
		Permissions: types.Permissions{
			DDL:                  true,
			ParamaterizedQueries: []string{"test_insert"},
		},
	}

	return &apipb.GetWalletRoleResponse{
		Name:        myRole.Name,
		Permissions: &apipb.GetWalletRoleResponsePerms{Ddl: myRole.Permissions.DDL, Queries: myRole.Permissions.ParamaterizedQueries},
	}, nil
}
