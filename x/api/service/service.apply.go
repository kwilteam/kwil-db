package service

import (
	"context"
	"fmt"
	"math/big"

	v0 "kwil/x/api/v0"
	"kwil/x/crypto"
)

func (s *Service) Apply(ctx context.Context, req *v0.ApplyRequest) (*v0.ApplyResponse, error) {

	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(req.Id))
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidSignature
	}

	id := applyID(req)

	if req.Id != id {
		return nil, ErrInvalidID
	}

	// TODO: validate payment??
	// Flat payment for now - based on diff later?

	// big int from fee
	amt, errb := big.NewInt(0).SetString(req.Fee, 10)
	if errb {
		return nil, fmt.Errorf("failed to parse fee: %s", req.Fee)
	}

	// validate that the fee is greater than or equal to 5
	if amt.Cmp(big.NewInt(5)) < 0 {
		return nil, ErrNotEnoughFunds
	}

	s.ds.Spend(req.From, amt)

	_ = getPlan(req.PlanId)

	// send plan

	return &v0.ApplyResponse{
		Success: true,
	}, nil
}

func getPlan(id string) *v0.PlanRequest {
	return &v0.PlanRequest{
		Id: id,
	}
}
