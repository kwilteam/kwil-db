package schemasvc

import (
	"context"
	schemapb "kwil/x/proto/schemasvc"
)

type service struct {
	schemapb.UnimplementedSchemaServiceServer
}

func New() schemapb.SchemaServiceServer {
	return &service{}
}

func (s *service) Plan(ctx context.Context, req *schemapb.PlanRequest) (*schemapb.PlanResponse, error) {
	return &schemapb.PlanResponse{}, nil
}

func (s *service) Apply(ctx context.Context, req *schemapb.ApplyRequest) (*schemapb.ApplyResponse, error) {
	return &schemapb.ApplyResponse{}, nil
}
