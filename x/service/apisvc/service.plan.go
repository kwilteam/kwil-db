package apisvc

import (
	"context"

	"kwil/x/crypto"
	"kwil/x/proto/apipb"
)

func (s *Service) Plan(ctx context.Context, req *apipb.PlanRequest) (*apipb.PlanResponse, error) {

	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(req.Id))
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidSignature
	}

	id := planID(req)

	if req.Id != id {
		return nil, ErrInvalidID
	}

	return &apipb.PlanResponse{}, nil
}
