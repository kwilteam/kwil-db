package txsvc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/engine"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	logger := s.log.With(zap.String("rpc", "GetSchema"), zap.String("dbid", req.Dbid))
	schema, err := s.engine.GetSchema(ctx, req.Dbid)
	if err != nil {
		logger.Error("failed to get schema", zap.Error(err))

		if err == engine.ErrDatasetNotFound {
			return nil, status.Error(codes.NotFound, "dataset not found")
		}

		return nil, status.Error(codes.Internal, "failed to get schema")
	}

	txSchema, err := convertSchemaFromEngine(schema)
	if err != nil {
		logger.Error("failed to convert schema", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to convert schema")
	}

	return &txpb.GetSchemaResponse{
		Schema: txSchema,
	}, nil
}
