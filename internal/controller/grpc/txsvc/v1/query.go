package txsvc

import (
	"context"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/internal/entity"
)

func (s *Service) Query(ctx context.Context, req *txpb.QueryRequest) (*txpb.QueryResponse, error) {
	bts, err := s.executor.Query(&entity.DBQuery{
		DBID:  req.Dbid,
		Query: req.Query,
	})
	if err != nil {
		return nil, err
	}

	return &txpb.QueryResponse{
		Result: bts,
	}, nil
}
