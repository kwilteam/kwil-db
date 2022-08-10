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

type HelloWorld struct {
	msg1 string
	msg2 string
}

type Events struct {
	HelloWorld HelloWorld
}

type EventFeed struct {
	log         zerolog.Logger
	ClientChain types.ClientChain
	EthClient   *ethclient.Client
	Topics      map[common.Hash]abi.Event
}

// Creates a new EventFeed
func New(conf *types.Config, ethClient *ethclient.Client) (*EventFeed, error) {
	logger := log.With().Str("module", "events").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	topics := GetTopics(conf)
	return &EventFeed{
		log:         logger,
		ClientChain: conf.ClientChain,
		EthClient:   ethClient,
		Topics:      topics,
	}, nil
}

func (e *EventFeed) Start(
	ctx context.Context,
) (chan map[string]interface{}, error) {
	e.log.Debug().Msg("starting event feed")
	defer e.log.Debug().Msg("event feed stopped")

	// This will return a channel that contains block height bigint values
	headers, err := e.listenForBlockHeaders(ctx)
	if err != nil {
		return nil, err
	}

	events := e.pullEvents(ctx, headers)

	return events, nil
}

// This function gets the list of topics
func (e *EventFeed) getTopicsForEvents() []common.Hash {
	topics := make([]common.Hash, len(e.Topics))
	for _, v := range e.Topics {
		topics = append(topics, v.ID)
	}
	return topics
}

func GetTopics(conf *types.Config) map[common.Hash]abi.Event {
	// First, get the ABI for the contract
	events := conf.ClientChain.GetContractABI().Events // Named this cAbi to avoid confusion with the abi.ABI type
	topics := make(map[common.Hash]abi.Event)

	for _, ev := range events {
		topics[ev.ID] = ev
	}
	return topics
}
