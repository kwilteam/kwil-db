package apisvc

import (
	"context"

	"kwil/x/proto/apipb"
)

func (s *Service) Connect(ctx context.Context, req *apipb.ConnectRequest) (*apipb.ConnectResponse, error) {
	return &apipb.ConnectResponse{Address: "0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D"}, nil
}
