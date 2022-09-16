package service

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
)

func (s *Service) CreateTable(ctx context.Context, req *v0.CreateTableRequest) (*v0.CreateTableResponse, error) {
	return &v0.CreateTableResponse{}, nil
}

func (s *Service) UpdateTable(ctx context.Context, req *v0.UpdateTableRequest) (*v0.UpdateTableResponse, error) {
	return &v0.UpdateTableResponse{}, nil
}

func (s *Service) ListTables(ctx context.Context, req *v0.ListTablesRequest) (*v0.ListTablesResponse, error) {
	return &v0.ListTablesResponse{}, nil
}

func (s *Service) GetTable(ctx context.Context, req *v0.GetTableRequest) (*v0.GetTableResponse, error) {
	return &v0.GetTableResponse{}, nil
}

func (s *Service) DeleteTable(ctx context.Context, req *v0.DeleteTableRequest) (*v0.DeleteTableResponse, error) {
	return &v0.DeleteTableResponse{}, nil
}
