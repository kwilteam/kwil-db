package txsvc

import (
	"context"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/pkg/engine/models"
	"kwil/pkg/utils/serialize"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	schema, err := s.executor.GetSchema(req.Dbid)
	if err != nil {
		return nil, err
	}

	convSchema, err := serialize.Convert[models.Dataset, txpb.Dataset](schema)
	if err != nil {
		return nil, err
	}

	return &txpb.GetSchemaResponse{
		Dataset: convSchema,
	}, nil
}
