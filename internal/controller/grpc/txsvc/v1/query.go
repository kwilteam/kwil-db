package txsvc

import (
	"context"
	"encoding/json"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (s *Service) Query(ctx context.Context, req *txpb.QueryRequest) (*txpb.QueryResponse, error) {
	result, err := s.engine.Query(ctx, req.Dbid, req.Query)
	if err != nil {
		return nil, err
	}

	bts, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &txpb.QueryResponse{
		Result: bts,
	}, nil
}
