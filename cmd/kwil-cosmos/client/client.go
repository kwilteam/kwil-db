package client

import (
	"context"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	ktypes "github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Client struct {
	log          zerolog.Logger
	cosmosClient cosmosclient.Client
	account      cosmosaccount.Account
}

const (
	addressPrefix string = "kaddr-"
)

func New(conf *ktypes.Config, accName string) (Client, error) {
	// Get logger
	logger := log.With().Str("module", "client").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	// Get a cosmos client
	c, err := cosmosclient.New(context.Background(), cosmosclient.WithAddressPrefix(addressPrefix))
	if err != nil {
		return Client{}, err
	}

	// Get cosmos account
	acc, err := c.Account(accName)
	if err != nil {
		return Client{}, err
	}

	return Client{
		log:          logger,
		cosmosClient: c,
		account:      acc,
	}, err
}
