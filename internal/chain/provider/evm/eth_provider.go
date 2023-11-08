package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	evmClient "github.com/kwilteam/kwil-db/internal/chain/provider/evm/client"
	escrowCtr "github.com/kwilteam/kwil-db/internal/chain/provider/evm/contracts/escrow"
	tokenCtr "github.com/kwilteam/kwil-db/internal/chain/provider/evm/contracts/token"

	"github.com/kwilteam/kwil-db/internal/chain/types"
)

type ethProvider struct {
	client *evmClient.EthClient
	token  *tokenCtr.TokenContract
	escrow *escrowCtr.EscrowContract
}

func New(endpoint string, chainCode types.ChainCode, tokenAddress string, escrowAddress string) (*ethProvider, error) {
	client, err := evmClient.New(endpoint, chainCode)
	if err != nil {
		return nil, err
	}

	token, err := tokenCtr.New(client, tokenAddress, chainCode.ToChainId())
	if err != nil {
		return nil, err
	}

	escrow, err := escrowCtr.New(client, escrowAddress, chainCode.ToChainId())
	if err != nil {
		return nil, err
	}

	return &ethProvider{
		client: client,
		token:  token,
		escrow: escrow,
	}, nil
}

// Provider Client methods

func (p *ethProvider) ChainCode() types.ChainCode {
	return p.client.ChainCode()
}

func (p *ethProvider) Endpoint() string {
	return p.client.Endpoint()
}

func (p *ethProvider) Close() error {
	return p.client.Close()
}

func (p *ethProvider) GetAccountNonce(ctx context.Context, addr string) (uint64, error) {
	return p.client.GetAccountNonce(ctx, addr)
}

func (p *ethProvider) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return p.client.SuggestGasPrice(ctx)
}

// Token Contract methods

func (p *ethProvider) Allowance(ctx context.Context, owner, spender string) (*big.Int, error) {
	return p.token.Allowance(ctx, owner, spender)
}

func (p *ethProvider) Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types.ApproveResponse, error) {
	return p.token.Approve(ctx, spender, amount, privateKey)
}

// Escrow Contract methods

func (p *ethProvider) Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error) {
	return p.escrow.Deposit(ctx, params, privateKey)
}
