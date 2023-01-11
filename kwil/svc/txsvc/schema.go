package txsvc

import (
	"context"
	"fmt"
	"kwil/x/proto/commonpb"
	"kwil/x/proto/txpb"
	"kwil/x/types/databases"
	"kwil/x/types/execution"
	"kwil/x/utils/serialize"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	db, err := s.dao.GetDatabase(ctx, &databases.DatabaseIdentifier{
		Name:  req.Database,
		Owner: req.Owner,
	})
	if err != nil {
		s.log.Sugar().Warnf("failed to get database", err)
		return nil, fmt.Errorf("failed to get database")
	}

	convDb, err := serialize.Convert[databases.Database, commonpb.Database](db)
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
	execs, err := s.executor.GetExecutables(ctx, &databases.DatabaseIdentifier{
		Name:  req.Database,
		Owner: req.Owner,
	})
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
	if err != nil {
		s.log.Sugar().Warnf("failed to convert executables", err)
		return nil, fmt.Errorf("failed to convert executables")
	}

	return &txpb.GetExecutablesResponse{
		Executables: convertedExecutables,
	}, nil
}
