package txsvc

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
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
