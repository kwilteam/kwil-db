package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

func (s *Service) ListDatabases(ctx context.Context, req *txpb.ListDatabasesRequest) (*txpb.ListDatabasesResponse, error) {
	dbs, err := s.engine.ListOwnedDatabases(ctx, req.Owner)
	if err != nil {
		return nil, err
	}

	return &txpb.ListDatabasesResponse{
		Databases: dbs,
	}, nil
}
