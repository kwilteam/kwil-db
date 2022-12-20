package service

import (
	"context"
	"fmt"
	"kwil/x/cfgx"
	"kwil/x/chain-client/builder"
	"kwil/x/chain-client/dto"
	"kwil/x/logx"
	"time"
)

type ChainClient interface {
	Listen(ctx context.Context, confirmed bool) (<-chan int64, error)
	GetLatestBlock(ctx context.Context, confirmed bool) (int64, error)
	GetDeposits(ctx context.Context, start, end int64) ([]*dto.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*dto.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *dto.ReturnFundsParams) (*dto.ReturnFundsResponse, error)
}

// chainClient implements the ChainClient interface
type chainClient struct {
	listener              dto.Listener
	escrow                dto.EscrowContract
	log                   logx.SugaredLogger
	maxBlockInterval      time.Duration
	requiredConfirmations int64
	chainCode             dto.ChainCode
}

func NewChainClient(cfg cfgx.Config, privateKey string) (ChainClient, error) {
	log := logx.New().Named("chain-client").Sugar()

	chainCode := dto.ChainCode(cfg.Int64("chain-code", 0))

	providerEndpoint := cfg.String("provider-endpoint")
	if providerEndpoint == "" {
		return nil, fmt.Errorf("provider endpoint is required")
	}

	contractAddress := cfg.String("contract-address")
	if contractAddress == "" {
		return nil, fmt.Errorf("contract address is required")
	}

	chainComponents, err := builder.NewChainBuilder().ChainCode(chainCode).ContractAddress(contractAddress).PrivateKey(privateKey).RPCProvider(providerEndpoint).Build()
	if err != nil {
		log.Fatalw("failed to build chain components", "error", err)
	}

	return &chainClient{
		listener:              chainComponents.Listener,
		escrow:                chainComponents.Escrow,
		log:                   log,
		maxBlockInterval:      time.Duration(cfg.Int64("reconnection-interval", 30)) * time.Second,
		requiredConfirmations: cfg.Int64("required-confirmations", 12),
		chainCode:             chainCode,
	}, nil
}

/*
func newChainSpecificClient(endpoint string, chainCode dto.ChainCode) (blockClient, error) {
	switch chainCode {
	case dto.ETHEREUM:
		return evmclient.New(endpoint, chainCode.ToChainId())
	case dto.GOERLI:
		return evmclient.New(endpoint, chainCode.ToChainId())
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainCode))
	}
}
*/
