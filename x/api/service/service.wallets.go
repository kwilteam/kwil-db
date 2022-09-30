package service

import (
	"context"

	types "kwil/pkg/types/db"
	v0 "kwil/x/api/v0"
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

func (s *Service) GetBalance(ctx context.Context, req *v0.GetBalanceRequest) (*v0.GetBalanceResponse, error) {

	bal, err := s.ds.GetBalance(req.Id)
	if err != nil {
		return nil, err
	}

	return &v0.GetBalanceResponse{
		Balance: bal.String(),
	}, nil
}
