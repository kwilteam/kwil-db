package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v0"
)

func (s *Service) Ping(ctx context.Context, req *txpb.PingRequest) (*txpb.PongResponse, error) {
	return &txpb.PongResponse{
		Message: "pong",
	}, nil
}
