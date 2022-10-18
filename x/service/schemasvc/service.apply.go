package schemasvc

import (
	"context"
	"kwil/x/proto/schemapb"
)

func (s *service) Apply(ctx context.Context, req *schemapb.ApplyRequest) (*schemapb.ApplyResponse, error) {
	return &schemapb.ApplyResponse{}, nil
}
