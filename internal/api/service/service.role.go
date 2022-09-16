package service

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
)

func (s *Service) CreateRole(ctx context.Context, req *v0.CreateRoleRequest) (*v0.CreateRoleResponse, error) {
	return &v0.CreateRoleResponse{}, nil
}

func (s *Service) UpdateRole(ctx context.Context, req *v0.UpdateRoleRequest) (*v0.UpdateRoleResponse, error) {
	return &v0.UpdateRoleResponse{}, nil
}

func (s *Service) ListRoles(ctx context.Context, req *v0.ListRolesRequest) (*v0.ListRolesResponse, error) {
	return &v0.ListRolesResponse{}, nil
}

func (s *Service) GetRole(ctx context.Context, req *v0.GetRoleRequest) (*v0.GetRoleResponse, error) {
	return &v0.GetRoleResponse{}, nil
}

func (s *Service) DeleteRole(ctx context.Context, req *v0.DeleteRoleRequest) (*v0.DeleteRoleResponse, error) {
	return &v0.DeleteRoleResponse{}, nil
}
