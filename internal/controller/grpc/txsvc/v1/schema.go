package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	schema, err := s.engine.GetSchema(ctx, req.Dbid)
	if err != nil {
		return nil, err
	}

	txSchema, err := convertSchemaFromEngine(schema)
	if err != nil {
		return nil, err
	}

	return &txpb.GetSchemaResponse{
		Schema: txSchema,
	}, nil
}
