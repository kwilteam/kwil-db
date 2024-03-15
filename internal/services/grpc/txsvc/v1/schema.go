package txsvc

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/core/rpc/client/user/grpc"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	logger := s.log.With(zap.String("rpc", "GetSchema"), zap.String("dbid", req.Dbid))
	schema, err := s.engine.GetSchema(req.Dbid)
	if err != nil {
		logger.Debug("failed to get schema", zap.Error(err))

		if errors.Is(err, execution.ErrDatasetNotFound) {
			return nil, status.Error(codes.NotFound, "dataset not found")
		}

		return nil, status.Error(codes.Unknown, "failed to get schema")
	}

	txSchema, err := grpc.ConvertSchemaToPB(schema)
	if err != nil {
		logger.Error("failed to convert schema", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to convert schema")
	}

	return &txpb.GetSchemaResponse{
		Schema: txSchema,
	}, nil
}
