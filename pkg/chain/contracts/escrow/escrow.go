package escrow

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/chain/contracts/escrow/evm"
	"kwil/pkg/chain/contracts/escrow/types"
	"kwil/pkg/chain/provider"
	chainTypes "kwil/pkg/chain/types"
	"kwil/pkg/log"
	"kwil/pkg/utils/retry"
)

type EscrowContract interface {
	GetDeposits(ctx context.Context, start, end int64, providerAddress string) ([]*types.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64, providerAddress string) ([]*types.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *types.ReturnFundsParams, privateKey *ecdsa.PrivateKey) (*types.ReturnFundsResponse, error)
	Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error)
	Balance(ctx context.Context, params *types.DepositBalanceParams) (*types.DepositBalanceResponse, error)
	TokenAddress() string
}

func New(chainProvider provider.ChainProvider, contractAddress string, opts ...EscrowOpts) (EscrowContract, error) {
	var ctr EscrowContract
	var err error
	switch chainProvider.ChainCode() {
	case chainTypes.ETHEREUM, chainTypes.GOERLI:

		ctr, err = evm.New(chainProvider, contractAddress)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainProvider.ChainCode()))
	}

	if err != nil {
		return nil, err
	}

	return newRetry(ctr, opts...), nil
}

type escrow struct {
	contract EscrowContract
	log      log.Logger
}

func newRetry(contract EscrowContract, opts ...EscrowOpts) EscrowContract {
	e := &escrow{
		contract: contract,
		log:      log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (r *escrow) retry(ctx context.Context, fn func() error) error {
	return retry.Retry(func() error {
		return fn()
	},
		retry.WithContext(ctx),
		retry.WithLogger(r.log),
		retry.WithMin(1),
		retry.WithMax(10),
		retry.WithFactor(2),
	)
}

func (r *escrow) GetDeposits(ctx context.Context, start, end int64, providerAddress string) ([]*types.DepositEvent, error) {
	var deposits []*types.DepositEvent
	err := r.retry(ctx, func() error {
		var err error
		deposits, err = r.contract.GetDeposits(ctx, start, end, providerAddress)
		return err
	})

	return deposits, err
}

func (r *escrow) GetWithdrawals(ctx context.Context, start, end int64, providerAddress string) ([]*types.WithdrawalConfirmationEvent, error) {
	var withdrawals []*types.WithdrawalConfirmationEvent
	err := r.retry(ctx, func() error {
		var err error
		withdrawals, err = r.contract.GetWithdrawals(ctx, start, end, providerAddress)
		return err
	})

	return withdrawals, err
}

func (r *escrow) ReturnFunds(ctx context.Context, params *types.ReturnFundsParams, privateKey *ecdsa.PrivateKey) (*types.ReturnFundsResponse, error) {
	var response *types.ReturnFundsResponse
	err := r.retry(ctx, func() error {
		var err error
		response, err = r.contract.ReturnFunds(ctx, params, privateKey)
		return err
	})

	return response, err
}

func (r *escrow) Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error) {
	var response *types.DepositResponse
	err := r.retry(ctx, func() error {
		var err error
		response, err = r.contract.Deposit(ctx, params, privateKey)
		return err
	})

	return response, err
}

func (r *escrow) Balance(ctx context.Context, params *types.DepositBalanceParams) (*types.DepositBalanceResponse, error) {
	var response *types.DepositBalanceResponse
	err := r.retry(ctx, func() error {
		var err error
		response, err = r.contract.Balance(ctx, params)
		return err
	})

	return response, err
}

func (r *escrow) TokenAddress() string {
	return r.contract.TokenAddress()
}
