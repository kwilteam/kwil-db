package apisvc

import (
	"context"

	"kwil/x/proto/apipb"
)

func (s *Service) PlanSchema(ctx context.Context, req *apipb.PlanSchemaRequest) (*apipb.PlanSchemaResponse, error) {
	return &apipb.PlanSchemaResponse{}, nil
}

func (s *Service) ApplySchema(ctx context.Context, req *apipb.ApplySchemaRequest) (*apipb.ApplySchemaResponse, error) {
	return &apipb.ApplySchemaResponse{}, nil
}
