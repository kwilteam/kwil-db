package events

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	w "github.com/kwilteam/kwil-db/internal/chain/utils"
	types "github.com/kwilteam/kwil-db/pkg/types/chain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
	This package is to separate the event intake from the event processing (i.e., handling deposits).
*/

type Config interface {
	GetContractABI() abi.ABI
	GetChainID() int
	GetDepositAddress() string
	GetReqConfirmations() int
	GetBufferSize() int
	GetBlockTimeout() int
	GetLowestHeight() int64
}

type EventFeed struct {
	log       *zerolog.Logger
	conf      Config
	EthClient *ethclient.Client
	Topics    map[common.Hash]abi.Event
	Wal       w.Wal
	ds        types.DepositStore
}

// New Creates a new EventFeed
func New(conf Config, ethClient *ethclient.Client, wal w.Wal, ds types.DepositStore) (*EventFeed, error) {
	logger := log.With().Str("module", "events").Int64("chainID", int64(conf.GetChainID())).Logger()
	topics := getTopics(conf)

	return &EventFeed{
		log:       &logger,
		conf:      conf,
		EthClient: ethClient,
		Topics:    topics,
		Wal:       wal,
		ds:        ds,
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

func getTopics(conf Config) map[common.Hash]abi.Event {
	// First, get the ABI for the contract
	events := conf.GetContractABI().Events // Named this cAbi to avoid confusion with the abi.ABI type
	topics := make(map[common.Hash]abi.Event)

	for _, ev := range events {
		topics[ev.ID] = ev
	}
	return topics
}
