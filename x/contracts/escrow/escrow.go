package escrow

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/x/chain"
	chainClient "kwil/x/chain/client"
	escrowTypes "kwil/x/types/contracts/escrow"

	"kwil/x/contracts/escrow/evm"
)

type EscrowContract interface {
	GetDeposits(ctx context.Context, start, end int64) ([]*escrowTypes.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*escrowTypes.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *escrowTypes.ReturnFundsParams) (*escrowTypes.ReturnFundsResponse, error)
	Deposit(ctx context.Context, params *escrowTypes.DepositParams) (*escrowTypes.DepositResponse, error)
	TokenAddress() string
}

func New(chainClient chainClient.ChainClient, privateKey *ecdsa.PrivateKey, address string) (EscrowContract, error) {
	switch chainClient.ChainCode() {
	case chain.ETHEREUM, chain.GOERLI:
		ethClient, err := chainClient.AsEthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get ethclient from chain client: %d", err)
		}

		return evm.New(ethClient, chainClient.ChainCode().ToChainId(), privateKey, address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainClient.ChainCode()))
	}
}
