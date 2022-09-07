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

<<<<<<< HEAD
type Config interface {
	GetContractABI() abi.ABI
	GetChainID() int
	GetDepositAddress() string
	GetReqConfirmations() int
	GetBufferSize() int
	GetBlockTimeout() int
	GetLowestHeight() int64
=======
type CosmClient interface {
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
}

type EventFeed struct {
	log       *zerolog.Logger
<<<<<<< HEAD
	conf      Config
=======
	Config    *types.Config
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	EthClient *ethclient.Client
	Topics    map[common.Hash]abi.Event
	Wal       types.Wal
	ds        types.DepositStore
}

// New Creates a new EventFeed
<<<<<<< HEAD
func New(conf Config, ethClient *ethclient.Client, wal types.Wal, ds types.DepositStore) (*EventFeed, error) {
	logger := log.With().Str("module", "events").Int64("chainID", int64(conf.GetChainID())).Logger()
=======
func New(conf *types.Config, ethClient *ethclient.Client, wal types.Wal, ds types.DepositStore) (*EventFeed, error) {
	logger := log.With().Str("module", "events").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	topics := getTopics(conf)

	return &EventFeed{
		log:       &logger,
<<<<<<< HEAD
		conf:      conf,
=======
		Config:    conf,
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
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

<<<<<<< HEAD
func getTopics(conf Config) map[common.Hash]abi.Event {
=======
func getTopics(conf *types.Config) map[common.Hash]abi.Event {
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	// First, get the ABI for the contract
	events := conf.GetContractABI().Events // Named this cAbi to avoid confusion with the abi.ABI type
	topics := make(map[common.Hash]abi.Event)

	for _, ev := range events {
		topics[ev.ID] = ev
	}
	return topics
}
