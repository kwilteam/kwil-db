package txsvc

import (
	"context"
	"kwil/x/proto/txpb"
)

func (s *Service) Ping(ctx context.Context, req *txpb.PingRequest) (*txpb.PongResponse, error) {
	return &txpb.PongResponse{
		Message: "pong",
	}, nil
}
