package txsvc

import (
	"context"
	txpb "kwil/api/protobuf/tx/v1"
)

func (s *Service) Ping(ctx context.Context, req *txpb.PingRequest) (*txpb.PingResponse, error) {
	return &txpb.PingResponse{
		Message: "pong",
	}, nil
}
