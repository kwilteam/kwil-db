package apisvc

import (
	"context"

	apipb "kwil/x/proto/apisvc"
)

func (s *Service) CreateTable(ctx context.Context, req *apipb.CreateTableRequest) (*apipb.CreateTableResponse, error) {
	return &apipb.CreateTableResponse{}, nil
}

func (s *Service) UpdateTable(ctx context.Context, req *apipb.UpdateTableRequest) (*apipb.UpdateTableResponse, error) {
	return &apipb.UpdateTableResponse{}, nil
}

func (s *Service) ListTables(ctx context.Context, req *apipb.ListTablesRequest) (*apipb.ListTablesResponse, error) {
	return &apipb.ListTablesResponse{}, nil
}

func (s *Service) GetTable(ctx context.Context, req *apipb.GetTableRequest) (*apipb.GetTableResponse, error) {
	return &apipb.GetTableResponse{}, nil
}

func (s *Service) DeleteTable(ctx context.Context, req *apipb.DeleteTableRequest) (*apipb.DeleteTableResponse, error) {
	return &apipb.DeleteTableResponse{}, nil
}
