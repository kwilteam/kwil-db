package service

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
)

func (s *Service) Connect(ctx context.Context, req *v0.ConnectRequest) (*v0.ConnectResponse, error) {
	return &v0.ConnectResponse{Address: "0x995d95245698212D4Af52c8031F614C3D3127994"}, nil
}
