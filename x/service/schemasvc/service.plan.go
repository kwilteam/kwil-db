package schemasvc

import (
	"context"
	schemapb "kwil/x/proto/schemasvc"
)

func (s *service) Plan(ctx context.Context, req *schemapb.PlanRequest) (*schemapb.PlanResponse, error) {
	return &schemapb.PlanResponse{}, nil
}
