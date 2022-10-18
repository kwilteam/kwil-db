package apisvc

import (
	"context"

	types "kwil/pkg/types/db"
	"kwil/x/proto/apipb"
)

func (s *Service) CreateRole(ctx context.Context, req *apipb.CreateRoleRequest) (*apipb.CreateRoleResponse, error) {
	return &apipb.CreateRoleResponse{}, nil
}

func (s *Service) UpdateRole(ctx context.Context, req *apipb.UpdateRoleRequest) (*apipb.UpdateRoleResponse, error) {
	return &apipb.UpdateRoleResponse{}, nil
}

func (s *Service) ListRoles(ctx context.Context, req *apipb.ListRolesRequest) (*apipb.ListRolesResponse, error) {

	// TODO: this needs to be implemented.
	return &apipb.ListRolesResponse{
		Roles: []string{"admin", "default"},
	}, nil
}

func (s *Service) GetRole(ctx context.Context, req *apipb.GetRoleRequest) (*apipb.GetRoleResponse, error) {

	myRole := types.Role{
		Name: "admin",
		Permissions: types.Permissions{
			DDL:                  true,
			ParamaterizedQueries: []string{"test_insert"},
		},
	}

	return &apipb.GetRoleResponse{
		Name:        myRole.Name,
		Permissions: &apipb.GetRoleResponsePerms{Ddl: myRole.Permissions.DDL, Queries: myRole.Permissions.ParamaterizedQueries},
	}, nil
}

func (s *Service) DeleteRole(ctx context.Context, req *apipb.DeleteRoleRequest) (*apipb.DeleteRoleResponse, error) {
	return &apipb.DeleteRoleResponse{}, nil
}
