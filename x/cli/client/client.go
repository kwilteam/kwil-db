package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	accountsDto "kwil/x/accounts/dto"
	chainClient "kwil/x/chain/client"
	chainClientDto "kwil/x/chain/client/dto"
	chainClientService "kwil/x/chain/client/service"
	"kwil/x/contracts/escrow"

	"kwil/x/proto/accountspb"
	"kwil/x/proto/pricingpb"
	"kwil/x/proto/txpb"
	txDto "kwil/x/transactions/dto"
	txUtils "kwil/x/transactions/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type Client struct {
	accounts accountspb.AccountServiceClient
	txs      txpb.TxServiceClient
	pricing  pricingpb.PricingServiceClient

	ChainCode        int64
	PrivateKey       *ecdsa.PrivateKey
	Address          *common.Address
	TokenAddress     *common.Address
	PoolAddress      *common.Address
	ValidatorAddress *common.Address
	escrow           escrow.EscrowContract
	//erc20            *abi.Erc20
	chainClient chainClient.ChainClient
}

func NewClient(cc *grpc.ClientConn, v *viper.Viper) (*Client, error) {
	chainCode := v.GetInt64("chain-code")
	fundingPool := common.HexToAddress(v.GetString("funding-pool"))
	nodeAddress := common.HexToAddress(v.GetString("node-address"))
	ethProvider := v.GetString("eth-provider")

	privateKey, err := crypto.HexToECDSA(v.GetString("private-key"))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	chnClient, err := chainClientService.NewChainClientExplicit(&chainClientDto.Config{
		ChainCode:             chainCode,
		Endpoint:              ethProvider,
		ReconnectionInterval:  30,
		RequiredConfirmations: 12,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	// escrow
	escrowCtr, err := escrow.New(chnClient, privateKey, fundingPool.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := common.HexToAddress(escrowCtr.TokenAddress())

	return &Client{
		accounts: accountspb.NewAccountServiceClient(cc),
		txs:      txpb.NewTxServiceClient(cc),
		pricing:  pricingpb.NewPricingServiceClient(cc),

		ChainCode:        chainCode,
		PrivateKey:       privateKey,
		Address:          &address,
		PoolAddress:      &fundingPool,
		ValidatorAddress: &nodeAddress,
		chainClient:      chnClient,
		escrow:           escrowCtr,
		TokenAddress:     &tokenAddress,
	}, nil
}

func (c *Client) GetAccount(ctx context.Context, address string) (*accountsDto.Account, error) {
	acc, err := c.accounts.GetAccount(ctx, &accountspb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	return &accountsDto.Account{
		Address: acc.Address,
		Balance: acc.Balance,
		Spent:   acc.Spent,
		Nonce:   acc.Nonce,
	}, nil
}

func (c *Client) EstimatePrice(ctx context.Context, tx *txDto.Transaction) (string, error) {
	// estimate cost
	fee, err := c.pricing.EstimateCost(ctx, &pricingpb.EstimateRequest{
		Tx: txUtils.TxToMsg(tx),
	})
	if err != nil {
		return "", err
	}

	return fee.Price, nil
}

func (c *Client) Broadcast(ctx context.Context, tx *txDto.Transaction) (*txDto.Response, error) {
	// broadcast
	broadcast, err := c.txs.Broadcast(ctx, &txpb.BroadcastRequest{
		Tx: txUtils.TxToMsg(tx),
	})
	if err != nil {
		return nil, err
	}

	return &txDto.Response{
		Hash: broadcast.Hash,
		Fee:  broadcast.Fee,
	}, nil
}
