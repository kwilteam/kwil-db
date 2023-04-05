package txsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0"
	txpb "kwil/api/protobuf/tx/v0"

	"kwil/pkg/databases"
	"kwil/pkg/databases/convert"
	"kwil/pkg/databases/executables"
	"kwil/pkg/utils/serialize"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	dbIdentifier, err := s.executor.GetDBIdentifier(req.Id)
	if err != nil {
		s.log.Sugar().Warnf("failed to get database identifier", err)
		return nil, fmt.Errorf("failed to get database identifier")
	}

	db, err := s.dao.GetDatabase(ctx, dbIdentifier)
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

func (s *Service) GetQueries(ctx context.Context, req *txpb.GetQueriesRequest) (*txpb.GetQueriesResponse, error) {
	execs, err := s.executor.GetQueries(req.Id)
	if err != nil {
		s.log.Sugar().Warnf("failed to get queries", err)
		return nil, fmt.Errorf("failed to get queries. ensure the database exists")
	}

	convertedExecutables := make([]*commonpb.QuerySignature, len(execs))
	for i, e := range execs {
		converted, err := serialize.Convert[executables.QuerySignature, commonpb.QuerySignature](e)
		if err != nil {
			s.log.Sugar().Warnf("failed to convert queries", err)
			return nil, fmt.Errorf("failed to convert queries")
		}
		convertedExecutables[i] = converted
	}

	return &txpb.GetQueriesResponse{
		Queries: convertedExecutables,
	}, nil
}
