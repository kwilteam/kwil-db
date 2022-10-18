package apisvc

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/x"
	"kwil/x/messaging/mx"
	"kwil/x/messaging/pub"
	request "kwil/x/request_service"
	"math/big"

	types "kwil/pkg/types/db"
	"kwil/x/crypto"
	"kwil/x/proto/apipb"
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

	amt, errb := big.NewInt(0).SetString(req.Fee, 10)
	if errb {
		return nil, err
	}

	// validate that the fee is enough
	if !s.validateBalances(&req.From, &req.Operation, &req.Crud, amt) {
		return nil, ErrNotEnoughFunds
	}

	err = s.ds.Spend(req.From, amt)
	if err != nil {
		return nil, fmt.Errorf("failed to set balance for %s: %w", req.From, err)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	request_id, err := doSubmitRequest(ctx, req, func(req *apipb.CreateDatabaseRequest) *DBRequest {
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

	amt, errb := big.NewInt(0).SetString(req.Fee, 10)
	if errb {
		return nil, err
	}

	// validate that the fee is enough
	if !s.validateBalances(&req.From, &req.Operation, &req.Crud, amt) {
		return nil, ErrNotEnoughFunds
	}

	err = s.ds.Spend(req.From, amt)
	if err != nil {
		return nil, fmt.Errorf("failed to set balance for %s: %w", req.From, err)
	}

	request_id, err := doSubmitRequest(ctx, req, func(req *apipb.UpdateDatabaseRequest) *DBRequest {
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

type createDatabaseRequestSerdes struct {
}

func GetDbRequestSerdes() mx.Serdes[*DBRequest] {
	panic("implement me")
}

func (c *createDatabaseRequestSerdes) Serialize(_ *DBRequest) (key []byte, value []byte, err error) {
	// TODO: get serialize logic from Bryan
	panic("not implemented")
}

func (c *createDatabaseRequestSerdes) Deserialize(_ []byte, _ []byte) (*DBRequest, error) {
	// TODO: get deserialize logic from Bryan
	panic("not implemented")
}

func doSubmitRequest[T any](ctx context.Context, req T, fn func(T) *DBRequest) (string, error) {
	emitter := x.Resolve[pub.Emitter[*DBRequest]](ctx, DATABASE_EMITTER_ALIAS)
	if emitter == nil {
		return "", fmt.Errorf("failed to resolve emitter %s", DATABASE_EMITTER_ALIAS)
	}

	if emitter == nil {
		return "", fmt.Errorf("failed to resolve emitter %s", DATABASE_EMITTER_ALIAS)
	}

	db_req := fn(req)
	a := emitter.Send(ctx, db_req)
	<-a.DoneCh() // blocking call

	requestManager := x.Resolve[request.Manager](ctx, request.MANAGER_ALIAS)
	if requestManager == nil {
		return "", fmt.Errorf("failed to resolve request manager %s", request.MANAGER_ALIAS)
	}

	info, err := requestManager.Create(ctx, db_req.IdempotentKey)
	if err == nil {
		return "", err
	}

	return info.ID(), nil
}
