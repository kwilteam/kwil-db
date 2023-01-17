package txsvc

import (
	"context"
	"fmt"
	"kwil/x/execution/clean"
	"kwil/x/execution/validation"
	"kwil/x/proto/commonpb"
	"kwil/x/proto/txpb"
	"kwil/x/types/databases"
	"kwil/x/utils/serialize"
)

func (s *Service) ValidateSchema(ctx context.Context, req *txpb.ValidateSchemaRequest) (*txpb.ValidateSchemaResponse, error) {
	res := &txpb.ValidateSchemaResponse{
		Valid: false,
	}

	// convert the database
	db, err := serialize.Convert[commonpb.Database, databases.Database](req.Schema)
	if err != nil {
		s.log.Sugar().Warnf("failed to convert database", err)
		return nil, fmt.Errorf("failed to convert request body")
	}

	// clean the database
	err = clean.CleanDatabase(db)
	if err != nil {
		s.log.Sugar().Warnf("failed to clean database", err)
		// we want to return this error message to the user
		res.Error = fmt.Errorf("error cleaning database: %w", err).Error()
		return res, nil
	}

	// validate the database
	err = validation.ValidateDatabase(db)
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
