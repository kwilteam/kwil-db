package apisvc

import (
	"context"
	"encoding/json"
	"fmt"
	types "kwil/pkg/types/db"
	"kwil/x/chain/crypto"
	"kwil/x/proto/apipb"
	"kwil/x/svcx/composer"
)

func (s *Service) CreateDatabase(ctx context.Context, req *apipb.CreateDatabaseRequest) (*apipb.CreateDatabaseResponse, error) {
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

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	request_id, err := doSubmitRequest(ctx, req, func(req *apipb.CreateDatabaseRequest) *composer.Message {
		return getCreateDbRequest(req)
	})

	_ = request_id // remove, not using compile error for now
	// need to update te response with the request id

	return &apipb.CreateDatabaseResponse{}, nil
}

func (s *Service) UpdateDatabase(ctx context.Context, req *apipb.UpdateDatabaseRequest) (*apipb.UpdateDatabaseResponse, error) {
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

	request_id, err := doSubmitRequest(ctx, req, func(req *apipb.UpdateDatabaseRequest) *composer.Message {
		return getUpdateDbRequest(req)
	})

	_ = request_id // remove, not using compile error for now
	// need to update te response with the request id

	return &apipb.UpdateDatabaseResponse{}, nil
}

func (s *Service) ListDatabases(_ context.Context, _ *apipb.ListDatabasesRequest) (*apipb.ListDatabasesResponse, error) {
	return &apipb.ListDatabasesResponse{}, nil
}

func (s *Service) GetDatabase(_ context.Context, _ *apipb.GetDatabaseRequest) (*apipb.GetDatabaseResponse, error) {
	return &apipb.GetDatabaseResponse{}, nil
}

func (s *Service) DeleteDatabase(_ context.Context, _ *apipb.DeleteDatabaseRequest) (*apipb.DeleteDatabaseResponse, error) {
	return &apipb.DeleteDatabaseResponse{}, nil
}

func (s *Service) PostQuery(_ context.Context, _ *apipb.PostQueryRequest) (*apipb.PostQueryResponse, error) {
	return &apipb.PostQueryResponse{
		Id:  "123",
		Msg: "success!",
	}, nil
}

func (s *Service) GetQueries(_ context.Context, _ *apipb.GetQueriesRequest) (*apipb.GetQueriesResponse, error) {
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
	var pqs apipb.GetQueriesResponse
	err = json.Unmarshal(b, &pqs)
	if err != nil {
		return nil, err
	}

	return &pqs, nil
}
