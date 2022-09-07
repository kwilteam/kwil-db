package client

import (
	"context"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

//This is used as both a cosmos client and eth client

type CosmosClient struct {
	Client  cosmosclient.Client
	conf    *types.Config
	log     zerolog.Logger
	address string
}

type EthClient struct {
}

//goland:noinspection GoUnusedExportedFunction
func NewCosmosClient(ctx context.Context, conf *types.Config) (CosmosClient, error) {
	logger := log.With().Str("location", "CosmosClient").Logger()
	account := CosmosClient{
		conf: conf,
		log:  logger,
	}

	cosmos, err := importWallet(ctx, conf)
	if err != nil {
		return account, err
	}

	an, err := cosmos.Account(conf.Wallets.Cosmos.KeyName)
	if err != nil {
		return account, err
	}

	account.address = an.Address(conf.Wallets.Cosmos.AddressPrefix)

	account.Client = cosmos
	return account, nil
}

func (c *CosmosClient) Transfer(amt int, toAddr string) error {
	tokenAmt := strconv.Itoa(amt) + "token"
	coins, err := sdk.ParseCoinNormalized(tokenAmt)
	if err != nil {
		return err
	}

	msg := &banktypes.MsgSend{
		FromAddress: c.address,
		ToAddress:   toAddr,
		Amount:      sdk.Coins{coins},
	}
	txResp, err := c.Client.BroadcastTx(c.conf.Wallets.Cosmos.KeyName, msg)
	if err != nil {
		c.log.Debug().Err(err).Msg("error broadcasting tx")
		return err
	}

	c.log.Debug().Interface("tx", txResp).Msg("tx broadcast")

	return nil
}
