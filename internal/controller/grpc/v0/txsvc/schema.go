package txsvc

import (
	"context"
	"fmt"
	"kwil/api/protobuf/kwil/common/v0/gen/go"
	_go2 "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/pkg/databases"
	"kwil/pkg/databases/convert"
	"kwil/pkg/types/execution"
	"kwil/pkg/utils/serialize"
)

func (s *Service) GetSchema(ctx context.Context, req *_go2.GetSchemaRequest) (*_go2.GetSchemaResponse, error) {
	return s.retrieveDatabaseSchema(ctx, &databases.DatabaseIdentifier{
		Owner: req.Owner,
		Name:  req.Name,
	})
}

func (s *Service) GetSchemaById(ctx context.Context, req *_go2.GetSchemaByIdRequest) (*_go2.GetSchemaResponse, error) {
	dbIdentifier, err := s.executor.GetDBIdentifier(req.Id)
	if err != nil {
		s.log.Sugar().Warnf("failed to get database identifier", err)
		return nil, fmt.Errorf("failed to get database identifier")
	}

	return s.retrieveDatabaseSchema(ctx, dbIdentifier)
}

func (s *Service) retrieveDatabaseSchema(ctx context.Context, database *databases.DatabaseIdentifier) (*_go2.GetSchemaResponse, error) {
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

	convDb, err := serialize.Convert[databases.Database[[]byte], _go.Database](byteDB)
	if err != nil {
		s.log.Sugar().Warnf("failed to convert database", err)
		return nil, fmt.Errorf("failed to return database metadata")
	}

	return &_go2.GetSchemaResponse{
		Database: convDb,
	}, nil
}

func (s *Service) ListDatabases(ctx context.Context, req *_go2.ListDatabasesRequest) (*_go2.ListDatabasesResponse, error) {
	dbs, err := s.dao.ListDatabasesByOwner(ctx, req.Owner)
	if err != nil {
		s.log.Sugar().Warnf("failed to list databases", err)
		return nil, fmt.Errorf("failed to list databases")
	}

	return &_go2.ListDatabasesResponse{
		Databases: dbs,
	}, nil
}

func (s *Service) GetExecutables(ctx context.Context, req *_go2.GetExecutablesRequest) (*_go2.GetExecutablesResponse, error) {
	id := databases.GenerateSchemaName(req.Owner, req.Name)
	return s.retrieveExecutables(id)
}

func (s *Service) GetExecutablesById(ctx context.Context, req *_go2.GetExecutablesByIdRequest) (*_go2.GetExecutablesResponse, error) {
	return s.retrieveExecutables(req.Id)
}

func (s *Service) retrieveExecutables(id string) (*_go2.GetExecutablesResponse, error) {
	execs, err := s.executor.GetExecutables(id)
	if err != nil {
		s.log.Sugar().Warnf("failed to get executables", err)
		return nil, fmt.Errorf("failed to get executables")
	}

	convertedExecutables := make([]*_go.Executable, len(execs))
	for i, e := range execs {
		converted, err := serialize.Convert[execution.Executable, _go.Executable](e)
		if err != nil {
			s.log.Sugar().Warnf("failed to convert executables", err)
			return nil, fmt.Errorf("failed to convert executables")
		}
		convertedExecutables[i] = converted
	}

	return &_go2.GetExecutablesResponse{
		Executables: convertedExecutables,
	}, nil
}
