package txsvc

import (
	"context"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
)

func (s *Service) Ping(ctx context.Context, req *txpb.PingRequest) (*txpb.PongResponse, error) {
	return &txpb.PongResponse{
		Message: "pong",
	}, nil
}
