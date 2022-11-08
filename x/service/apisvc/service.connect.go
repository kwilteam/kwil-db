package apisvc

import (
	"context"

	"kwil/x/proto/apipb"
)

func (s *Service) Connect(ctx context.Context, req *apipb.ConnectRequest) (*apipb.ConnectResponse, error) {
	return &apipb.ConnectResponse{Address: s.ds.Address()}, nil
}
