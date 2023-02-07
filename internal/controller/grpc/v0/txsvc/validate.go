package txsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/kwil/common/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/pkg/execution/validator"
	"kwil/pkg/types/databases"
	"kwil/pkg/types/databases/clean"
	"kwil/pkg/types/databases/convert"
	"kwil/pkg/utils/serialize"
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
	id := databases.GenerateSchemaName(db.Owner, db.Name)

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
