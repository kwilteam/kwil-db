package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

func (s *Service) Ping(ctx context.Context, req *txpb.PingRequest) (*txpb.PingResponse, error) {
	return &txpb.PingResponse{
		Message: s.chainID, // yeah, this is lazy, but it's a helpful HELLO
	}, nil
}
