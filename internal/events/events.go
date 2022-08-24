package events

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
	This package is to separate the event intake from the event processing (i.e., handling deposits).
*/

type CosmClient interface {
}

type EventFeed struct {
	log        *zerolog.Logger
	Config     *types.Config
	EthClient  *ethclient.Client
	CosmClient CosmClient
	Topics     map[common.Hash]abi.Event
	Wal        types.Wal
	ds         types.DepositStore
}

// New Creates a new EventFeed
func New(conf *types.Config, ethClient *ethclient.Client, cosmClient CosmClient, wal types.Wal, ds types.DepositStore) (*EventFeed, error) {
	logger := log.With().Str("module", "events").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	topics := getTopics(conf)

	return &EventFeed{
		log:        &logger,
		Config:     conf,
		EthClient:  ethClient,
		CosmClient: cosmClient,
		Topics:     topics,
		Wal:        wal,
		ds:         ds,
	}, nil
}

func (ef *EventFeed) Listen(
	ctx context.Context,
) error {
	ef.log.Debug().Msg("starting event feed")

	headers, err := ef.listenForBlockHeaders(ctx)
	if err != nil {
		return err
	}
	ef.processBlocks(ctx, headers)

	return nil
}

// This function gets the list of topics
func (ef *EventFeed) getTopicsForEvents() []common.Hash {
	topics := make([]common.Hash, len(ef.Topics))
	for _, v := range ef.Topics {
		topics = append(topics, v.ID)
	}
	return topics
}

func getTopics(conf *types.Config) map[common.Hash]abi.Event {
	// First, get the ABI for the contract
	events := conf.ClientChain.GetContractABI().Events // Named this cAbi to avoid confusion with the abi.ABI type
	topics := make(map[common.Hash]abi.Event)

	for _, ev := range events {
		topics[ev.ID] = ev
	}
	return topics
}
