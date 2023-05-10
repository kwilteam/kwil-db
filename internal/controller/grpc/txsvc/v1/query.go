package txsvc

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
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
