package apisvc

import (
	"context"

	types "kwil/pkg/types/db"
	"kwil/x/crypto"
	"kwil/x/proto/apipb"
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

func (s *Service) GetWithdrawalsForWallet(ctx context.Context, req *apipb.GetWithdrawalsRequest) (*apipb.GetWithdrawalsResponse, error) {
	wdr, err := s.ds.GetWithdrawalsForWallet(req.Wallet)
	if err != nil {
		return nil, err
	}

	// TODO: Im sure there is a better way to do this...

	var wds []*apipb.Withdrawal
	for _, wd := range wdr {
		wds = append(wds, &apipb.Withdrawal{
			Tx:            wd.Tx,
			Amount:        wd.Amount,
			Fee:           wd.Fee,
			CorrelationId: wd.Cid,
			Expiration:    wd.Expiration,
		})
	}

	return &apipb.GetWithdrawalsResponse{
		Withdrawals: wds,
	}, nil
}

func (s *Service) GetBalance(ctx context.Context, req *apipb.GetBalanceRequest) (*apipb.GetBalanceResponse, error) {

	bal, sp, err := s.ds.GetBalanceAndSpent(req.Wallet)
	if err != nil {
		return nil, err
	}

	return &apipb.GetBalanceResponse{
		Balance: bal,
		Spent:   sp,
	}, nil
}

func (s *Service) ReturnFunds(ctx context.Context, req *apipb.ReturnFundsRequest) (*apipb.ReturnFundsResponse, error) {

	// THIS SHOULD NOT YET BE USED IN PRODUCTION WITH REAL FUNDS

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
	wdr, err := s.ds.Withdraw(ctx, req.From, req.Amount)
	if err != nil {
		return nil, err
	}

	return &apipb.ReturnFundsResponse{
		Tx:            wdr.Tx,
		Amount:        wdr.Amount,
		Fee:           wdr.Fee,
		CorrelationId: wdr.Cid,
		Expiration:    wdr.Expiration,
	}, nil
}
