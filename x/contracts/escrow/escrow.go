package escrow

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	chainClient "kwil/pkg/chain/client"
	"kwil/pkg/chain/types"
	"kwil/x/contracts/escrow/evm"
	escrowTypes "kwil/x/types/contracts/escrow"
)

type EscrowContract interface {
	GetDeposits(ctx context.Context, start, end int64) ([]*escrowTypes.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*escrowTypes.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *escrowTypes.ReturnFundsParams) (*escrowTypes.ReturnFundsResponse, error)
	Deposit(ctx context.Context, params *escrowTypes.DepositParams) (*escrowTypes.DepositResponse, error)
	Balance(ctx context.Context, params *escrowTypes.DepositBalanceParams) (*escrowTypes.DepositBalanceResponse, error)
	TokenAddress() string
}

func New(chainClient chainClient.ChainClient, privateKey *ecdsa.PrivateKey, address string) (EscrowContract, error) {
	switch chainClient.ChainCode() {
	case types.ETHEREUM, types.GOERLI:
		ethClient, err := chainClient.AsEthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get ethclient from chain client: %d", err)
		}

		return evm.New(ethClient, chainClient.ChainCode().ToChainId(), privateKey, address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainClient.ChainCode()))
	}
}
