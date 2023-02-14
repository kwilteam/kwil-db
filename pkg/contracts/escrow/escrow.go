package escrow

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	chainClient "kwil/pkg/chain/client"
	chainTypes "kwil/pkg/chain/types"
	"kwil/pkg/contracts/escrow/evm"
	types "kwil/pkg/contracts/escrow/types"
)

type EscrowContract interface {
	GetDeposits(ctx context.Context, start, end int64, providerAddress string) ([]*types.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64, providerAddress string) ([]*types.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *types.ReturnFundsParams, privateKey *ecdsa.PrivateKey) (*types.ReturnFundsResponse, error)
	Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error)
	Balance(ctx context.Context, params *types.DepositBalanceParams) (*types.DepositBalanceResponse, error)
	TokenAddress() string
}

func New(chainClient chainClient.ChainClient, contractAddress string) (EscrowContract, error) {
	switch chainClient.ChainCode() {
	case chainTypes.ETHEREUM, chainTypes.GOERLI:
		ethClient, err := chainClient.AsEthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get ethclient from chain client: %d", err)
		}

		return evm.New(ethClient, chainClient.ChainCode().ToChainId(), contractAddress)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainClient.ChainCode()))
	}
}
