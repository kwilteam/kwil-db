package apisvc

import (
	"context"

	apipb "kwil/x/proto/apisvc"
)

func (s *Service) PlanSchema(ctx context.Context, req *apipb.PlanSchemaRequest) (*apipb.PlanSchemaResponse, error) {
	return &apipb.PlanSchemaResponse{}, nil
}

func (s *Service) ApplySchema(ctx context.Context, req *apipb.ApplySchemaRequest) (*apipb.ApplySchemaResponse, error) {
	return &apipb.ApplySchemaResponse{}, nil
}
