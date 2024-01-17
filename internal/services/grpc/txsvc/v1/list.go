package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

func (s *Service) ListDatabases(ctx context.Context, req *txpb.ListDatabasesRequest) (*txpb.ListDatabasesResponse, error) {
	dbs, err := s.engine.ListDatasets(ctx, req.Owner)
	if err != nil {
		return nil, err
	}

	pbDatasets := make([]*txpb.DatasetInfo, len(dbs))
	for i, db := range dbs {
		pbDatasets[i] = &txpb.DatasetInfo{
			Dbid:  db.DBID,
			Name:  db.Name,
			Owner: db.Owner,
		}
	}

	return &txpb.ListDatabasesResponse{
		Databases: pbDatasets,
	}, nil
}
