package apisvc

import (
	"context"
	"fmt"
	"math/big"

	"kwil/x/crypto"
	"kwil/x/proto/apipb"
)

func (s *Service) Apply(ctx context.Context, req *apipb.ApplyRequest) (*apipb.ApplyResponse, error) {

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
	amt, ok := big.NewInt(0).SetString(req.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse fee: %s", req.Fee)
	}

	// validate that the fee is greater than or equal to 5
	if amt.Cmp(big.NewInt(5)) < 0 {
		return nil, ErrNotEnoughFunds
	}

	s.ds.Spend(req.From, amt)

	_ = getPlan(req.PlanId)

	// send plan

	return &apipb.ApplyResponse{
		Success: true,
	}, nil
}

func getPlan(id string) *apipb.PlanRequest {
	return &apipb.PlanRequest{
		Id: id,
	}
}
