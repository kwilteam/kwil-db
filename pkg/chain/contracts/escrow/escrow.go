package escrow

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/chain/contracts/escrow/evm"
	"kwil/pkg/chain/contracts/escrow/types"
	"kwil/pkg/chain/provider"
	chainTypes "kwil/pkg/chain/types"
)

type EscrowContract interface {
	GetDeposits(ctx context.Context, start, end int64, providerAddress string) ([]*types.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64, providerAddress string) ([]*types.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *types.ReturnFundsParams, privateKey *ecdsa.PrivateKey) (*types.ReturnFundsResponse, error)
	Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error)
	Balance(ctx context.Context, params *types.DepositBalanceParams) (*types.DepositBalanceResponse, error)
	TokenAddress() string
}

func New(chainProvider provider.ChainProvider, contractAddress string) (EscrowContract, error) {
	switch chainProvider.ChainCode() {
	case chainTypes.ETHEREUM, chainTypes.GOERLI:

		return evm.New(chainProvider, contractAddress)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainProvider.ChainCode()))
	}
}
