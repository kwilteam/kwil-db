package service

import (
	"context"
	"encoding/json"
	"fmt"

	types "kwil/pkg/types/db"
	v0 "kwil/x/api/v0"
	"kwil/x/chain/crypto"
)

func (s *Service) CreateDatabase(ctx context.Context, req *v0.CreateDatabaseRequest) (*v0.CreateDatabaseResponse, error) {
	if req.Operation != 0 {
		return nil, ErrIncorrectOperation
	}

	if req.Crud != 0 {
		return nil, ErrIncorrectCrud
	}

	if req.Id != createDatabaseID(req.From, req.Name, req.Fee) {
		return nil, ErrInvalidID
	}

	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(req.Id))
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidSignature
	}

	amt, err := s.validateBalances(&req.From, &req.Operation, &req.Crud, &req.Fee)
	if err != nil {
		return nil, err
	}

	err = s.ds.SetBalance(req.From, amt)
	if err != nil {
		return nil, fmt.Errorf("failed to set balance for %s: %w", req.From, err)
	}

	return &v0.CreateDatabaseResponse{}, nil
}

func (s *Service) UpdateDatabase(ctx context.Context, req *v0.UpdateDatabaseRequest) (*v0.UpdateDatabaseResponse, error) {
	if req.Id != updateDatabaseID(req) {
		return nil, ErrInvalidID
	}

	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(req.Id))
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidSignature
	}

	amt, err := s.validateBalances(&req.From, &req.Operation, &req.Crud, &req.Fee)
	if err != nil {
		return nil, err
	}

	err = s.ds.SetBalance(req.From, amt)
	if err != nil {
		return nil, fmt.Errorf("failed to set balance for %s: %w", req.From, err)
	}

	// Forward message to Kafka
	return &v0.UpdateDatabaseResponse{}, nil
}

func (s *Service) ListDatabases(ctx context.Context, req *v0.ListDatabasesRequest) (*v0.ListDatabasesResponse, error) {
	return &v0.ListDatabasesResponse{}, nil
}

func (s *Service) GetDatabase(ctx context.Context, req *v0.GetDatabaseRequest) (*v0.GetDatabaseResponse, error) {
	return &v0.GetDatabaseResponse{}, nil
}

func (s *Service) DeleteDatabase(ctx context.Context, req *v0.DeleteDatabaseRequest) (*v0.DeleteDatabaseResponse, error) {
	return &v0.DeleteDatabaseResponse{}, nil
}

func (s *Service) PostQuery(ctx context.Context, req *v0.PostQueryRequest) (*v0.PostQueryResponse, error) {
	return &v0.PostQueryResponse{
		Id:  "123",
		Msg: "success!",
	}, nil
}

func (s *Service) GetQueries(ctx context.Context, req *v0.GetQueriesRequest) (*v0.GetQueriesResponse, error) {
	// TODO: Implement
	// returning some mock data right now

	q := struct {
		Queries []*types.ParameterizedQuery `json:"queries"`
	}{
		Queries: []*types.ParameterizedQuery{
			{
				Name:  "test_insert",
				Query: "Insert into...",
				Parameters: []types.Parameter{
					{
						Name: "name",
						Type: "string",
					},
					{
						Name: "age",
						Type: "int32",
					},
				},
			},
			{
				Name:  "test2",
				Query: "Select * from...",
				Parameters: []types.Parameter{
					{
						Name: "name",
						Type: "string",
					},
					{
						Name: "height",
						Type: "int32",
					},
				},
			},
		},
	}

	// marshalling then unmarshalling to get the correct type.  Is this the best way to do this?
	// convert to bytes
	b, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	// unmarshal to proto
	var pqs v0.GetQueriesResponse
	err = json.Unmarshal(b, &pqs)
	if err != nil {
		return nil, err
	}

	return &pqs, nil
}
