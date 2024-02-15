package txsvc

import (
	"context"
	"encoding/json"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Query(ctx context.Context, req *txpb.QueryRequest) (*txpb.QueryResponse, error) {
	tx, err := s.db.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result, err := s.engine.Query(ctx, tx, req.Dbid, req.Query)
	if err != nil {
		// We don't know for sure that it's an invalid argument, but an invalid
		// user-provided query isn't an internal server error.
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	bts, err := json.Marshal(result.Map()) // marshalling the map is less efficient, but necessary for backwards compatibility
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal call result")
	}

	return &txpb.QueryResponse{
		Result: bts,
	}, nil
}
