package txsvc

import (
	"context"
	"encoding/json"

	sql "github.com/kwilteam/kwil-db/common/sql"
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

	ctxExec, cancel := context.WithTimeout(ctx, s.readTxTimeout)
	defer cancel()

	result, err := s.engine.Execute(ctxExec, tx, req.Dbid, req.Query, nil)
	if err != nil {
		// We don't know for sure that it's an invalid argument, but an invalid
		// user-provided query isn't an internal server error.
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	bts, err := json.Marshal(resultMap(result)) // marshalling the map is less efficient, but necessary for backwards compatibility
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal call result")
	}

	return &txpb.QueryResponse{
		Result: bts,
	}, nil
}

func resultMap(r *sql.ResultSet) []map[string]any {
	m := make([]map[string]any, len(r.Rows))
	for i, row := range r.Rows {
		m2 := make(map[string]any)
		for j, col := range row {
			m2[r.Columns[j]] = col
		}

		m[i] = m2
	}

	return m
}
