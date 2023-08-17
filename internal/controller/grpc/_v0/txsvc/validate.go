package txsvc

import (
	"context"
	"fmt"

	commonpb "github.com/kwilteam/kwil-db/api/protobuf/common/v0"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v0"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/clean"
	"github.com/kwilteam/kwil-db/pkg/databases/convert"
	"github.com/kwilteam/kwil-db/pkg/databases/validator"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
)

func (s *Service) ValidateSchema(ctx context.Context, req *txpb.ValidateSchemaRequest) (*txpb.ValidateSchemaResponse, error) {
	res := &txpb.ValidateSchemaResponse{
		Valid: false,
	}

	// convert the database
	db, err := serialize.Convert[commonpb.Database, databases.Database[[]byte]](req.Schema)
	if err != nil {
		s.log.Sugar().Warnf("failed to convert database", err)
		return nil, fmt.Errorf("failed to convert request body")
	}

	// clean the database
	clean.Clean(db)

	// convert
	anyDB, err := convert.Bytes.DatabaseToKwilAny(db)
	if err != nil {
		s.log.Sugar().Warnf("failed to convert database to bytes", err)
		return nil, fmt.Errorf("failed to convert database to bytes")
	}

	// validate the database
	vdr := validator.Validator{}
	err = vdr.Validate(anyDB)
	if err != nil {
		s.log.Sugar().Warnf("failed to validate database", err)
		// we want to return this error message to the user
		res.Error = fmt.Errorf("error validating database: %w", err).Error()
		return res, nil
	}

	// generate id
	id := databases.GenerateSchemaId(db.Owner, db.Name)

	// check if the database already exists
	// this can be done by checking if the database id already exists in the executor
	_, err = s.executor.GetDBIdentifier(id)
	if err == nil {
		// if the database already exists, return an error
		s.log.Sugar().Warnf("database already exists", err)
		res.Error = fmt.Errorf("database already exists").Error()
		return res, nil
	}

	return &txpb.ValidateSchemaResponse{
		Valid: true,
	}, nil
}
