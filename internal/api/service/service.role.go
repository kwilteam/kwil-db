package service

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/v0"
	types "github.com/kwilteam/kwil-db/pkg/types/db"
)

func (s *Service) CreateRole(ctx context.Context, req *v0.CreateRoleRequest) (*v0.CreateRoleResponse, error) {
	return &v0.CreateRoleResponse{}, nil
}

func (s *Service) UpdateRole(ctx context.Context, req *v0.UpdateRoleRequest) (*v0.UpdateRoleResponse, error) {
	return &v0.UpdateRoleResponse{}, nil
}

func (s *Service) ListRoles(ctx context.Context, req *v0.ListRolesRequest) (*v0.ListRolesResponse, error) {

	// TODO: this needs to be implemented.
	return &v0.ListRolesResponse{
		Roles: []string{"admin", "default"},
	}, nil
}

func (s *Service) GetRole(ctx context.Context, req *v0.GetRoleRequest) (*v0.GetRoleResponse, error) {

	myRole := types.Role{
		Name: "admin",
		Permissions: types.Permissions{
			DDL:                  true,
			ParamaterizedQueries: []string{"test_insert"},
		},
	}

	return &v0.GetRoleResponse{
		Name:        myRole.Name,
		Permissions: &v0.GetRoleResponsePerms{Ddl: myRole.Permissions.DDL, Queries: myRole.Permissions.ParamaterizedQueries},
	}, nil
}

func (s *Service) DeleteRole(ctx context.Context, req *v0.DeleteRoleRequest) (*v0.DeleteRoleResponse, error) {
	return &v0.DeleteRoleResponse{}, nil
}
