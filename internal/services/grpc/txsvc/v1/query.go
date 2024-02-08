package txsvc

import (
	"context"
	"encoding/json"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

func (s *Service) Query(ctx context.Context, req *txpb.QueryRequest) (*txpb.QueryResponse, error) {
	tx, err := s.db.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result, err := s.engine.Query(ctx, tx, req.Dbid, req.Query)
	if err != nil {
		return nil, err
	}

	bts, err := json.Marshal(result.Map()) // marshalling the map is less efficient, but necessary for backwards compatibility
	if err != nil {
		return nil, err
	}

	return &txpb.QueryResponse{
		Result: bts,
	}, nil
}
