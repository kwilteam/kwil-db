package schemasvc

import (
	"context"
	"kwil/x/proto/schemapb"
)

func (s *service) Plan(ctx context.Context, req *schemapb.PlanRequest) (*schemapb.PlanResponse, error) {
	return &schemapb.PlanResponse{}, nil
}
