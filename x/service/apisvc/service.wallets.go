package apisvc

import (
	"context"
	"math/big"

	types "kwil/pkg/types/db"
	"kwil/x/crypto"
	"kwil/x/proto/apipb"

	"github.com/ethereum/go-ethereum/common"
)

func (s *Service) GetWalletRole(ctx context.Context, req *apipb.GetWalletRoleRequest) (*apipb.GetWalletRoleResponse, error) {

	myRole := types.Role{
		Name: "admin",
		Permissions: types.Permissions{
			DDL:                  true,
			ParamaterizedQueries: []string{"test_insert"},
		},
	}

	return &apipb.GetWalletRoleResponse{
		Name:        myRole.Name,
		Permissions: &apipb.GetWalletRoleResponsePerms{Ddl: myRole.Permissions.DDL, Queries: myRole.Permissions.ParamaterizedQueries},
	}, nil
}

func (s *Service) GetBalance(ctx context.Context, req *apipb.GetBalanceRequest) (*apipb.GetBalanceResponse, error) {

	bal, err := s.ds.GetBalance(req.Id)
	if err != nil {
		return nil, err
	}

	return &apipb.GetBalanceResponse{
		Balance: bal.String(),
	}, nil
}

func (s *Service) ReturnFunds(ctx context.Context, req *apipb.ReturnFundsRequest) (*apipb.ReturnFundsResponse, error) {

	// THIS SHOULD NOT YET BE USED IN PRODUCTION

	// reconstruct id
	// id for return funds is generated from amount, nonce, and address (from)

	id := createFundsReturnID(req.Amount, req.Nonce, req.From)

	if id != req.Id {
		return nil, ErrInvalidID
	}

	// check to make sure the the ID is signed
	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(req.Id))
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidSignature
	}

	// now validate that the caller has enough funds to return
	bal, err := s.ds.GetBalance(req.From)
	if err != nil {
		return nil, err
	}

	// convert req.Amount to big.Int
	amt, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return nil, ErrInvalidAmount
	}

	// check to make sure the amount is <= the balance
	if amt.Cmp(bal) > 0 {
		return nil, ErrNotEnoughFunds
	}

	// now we need to get the amount they have spent, since we will cash it out in the smart contract
	_, err = s.ds.GetSpent(req.From)
	if err != nil {
		return nil, err
	}

	// now we call the smart contract to return the funds
	// convert req.From to common.Address
	_ = common.HexToAddress(req.From)

	/*
		_, err = s.cc.ReturnFunds(ctx, from, amt, spent) // we need to build tracing here
		if err != nil {
			return nil, err
		}
	*/

	return &apipb.ReturnFundsResponse{}, nil
}
