package apisvc

import (
	"context"

	"kwil/x/proto/apipb"
)

func (s *Service) Connect(ctx context.Context, req *apipb.ConnectRequest) (*apipb.ConnectResponse, error) {
	return &apipb.ConnectResponse{Address: "0x995d95245698212D4Af52c8031F614C3D3127994"}, nil
}
