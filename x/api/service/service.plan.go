package service

import (
	"context"

	v0 "kwil/x/api/v0"
	"kwil/x/crypto"
)

func (s *Service) Plan(ctx context.Context, req *v0.PlanRequest) (*v0.PlanResponse, error) {

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

	return &v0.PlanResponse{}, nil
}
