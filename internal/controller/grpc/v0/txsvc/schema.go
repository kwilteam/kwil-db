package txsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0/gen/go"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/x/types/databases"
	"kwil/x/types/databases/convert"
	"kwil/x/types/execution"
	"kwil/x/utils/serialize"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	return s.retrieveDatabaseSchema(ctx, &databases.DatabaseIdentifier{
		Owner: req.Owner,
		Name:  req.Name,
	})
}

func (s *Service) GetSchemaById(ctx context.Context, req *txpb.GetSchemaByIdRequest) (*txpb.GetSchemaResponse, error) {
	dbIdentifier, err := s.executor.GetDBIdentifier(req.Id)
	if err != nil {
		s.log.Sugar().Warnf("failed to get database identifier", err)
		return nil, fmt.Errorf("failed to get database identifier")
	}

	return s.retrieveDatabaseSchema(ctx, dbIdentifier)
}

func (s *Service) retrieveDatabaseSchema(ctx context.Context, database *databases.DatabaseIdentifier) (*txpb.GetSchemaResponse, error) {
	db, err := s.dao.GetDatabase(ctx, database)
	if err != nil {
		s.log.Sugar().Warnf("failed to get database", err)
		return nil, fmt.Errorf("failed to get database")
	}

	byteDB, err := convert.KwilAny.DatabaseToBytes(db)
	if err != nil {
		s.log.Sugar().Warnf("failed to convert database to bytes", err)
		return nil, fmt.Errorf("failed to return database metadata")
	}

	convDb, err := serialize.Convert[databases.Database[[]byte], commonpb.Database](byteDB)
	if err != nil {
		s.log.Sugar().Warnf("failed to convert database", err)
		return nil, fmt.Errorf("failed to return database metadata")
	}

	return &txpb.GetSchemaResponse{
		Database: convDb,
	}, nil
}

func (s *Service) ListDatabases(ctx context.Context, req *txpb.ListDatabasesRequest) (*txpb.ListDatabasesResponse, error) {
	dbs, err := s.dao.ListDatabasesByOwner(ctx, req.Owner)
	if err != nil {
		s.log.Sugar().Warnf("failed to list databases", err)
		return nil, fmt.Errorf("failed to list databases")
	}

	return &txpb.ListDatabasesResponse{
		Databases: dbs,
	}, nil
}

func (s *Service) GetExecutables(ctx context.Context, req *txpb.GetExecutablesRequest) (*txpb.GetExecutablesResponse, error) {
	id := databases.GenerateSchemaName(req.Owner, req.Name)
	return s.retrieveExecutables(id)
}

func (s *Service) GetExecutablesById(ctx context.Context, req *txpb.GetExecutablesByIdRequest) (*txpb.GetExecutablesResponse, error) {
	return s.retrieveExecutables(req.Id)
}

func (s *Service) retrieveExecutables(id string) (*txpb.GetExecutablesResponse, error) {
	execs, err := s.executor.GetExecutables(id)
	if err != nil {
		s.log.Sugar().Warnf("failed to get executables", err)
		return nil, fmt.Errorf("failed to get executables")
	}

	convertedExecutables := make([]*commonpb.Executable, len(execs))
	for i, e := range execs {
		converted, err := serialize.Convert[execution.Executable, commonpb.Executable](e)
		if err != nil {
			s.log.Sugar().Warnf("failed to convert executables", err)
			return nil, fmt.Errorf("failed to convert executables")
		}
		convertedExecutables[i] = converted
	}

	return &txpb.GetExecutablesResponse{
		Executables: convertedExecutables,
	}, nil
}
