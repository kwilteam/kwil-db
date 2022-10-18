package schemasvc

import (
	"context"
	schemapb "kwil/x/proto/schemasvc"
)

func (s *service) Apply(ctx context.Context, req *schemapb.ApplyRequest) (*schemapb.ApplyResponse, error) {
	return &schemapb.ApplyResponse{}, nil
}
