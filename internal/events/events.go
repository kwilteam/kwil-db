package events

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/big"
	"sync"
)

/*
	This package is to separate the event intake from the event processing (i.e., handling deposits).
*/

type DepositStore interface {
	Deposit(amount *big.Int, addr string, tx []byte, height *big.Int) error
	GetBalance(addr string) (*big.Int, error)
	CommitBlock(height *big.Int) error
	GetLastHeight() (*big.Int, error)
	SetLastHeight(height *big.Int) error
}

type EventFeed struct {
	log       zerolog.Logger
	Config    *types.Config
	EthClient *ethclient.Client
	Topics    map[common.Hash]abi.Event
	Wal       types.Wal
	ds        DepositStore
	mu        sync.Mutex
}

const walPath = ".wal"

// Creates a new EventFeed
func New(conf *types.Config, ethClient *ethclient.Client, wal types.Wal, ds DepositStore) (*EventFeed, error) {
	logger := log.With().Str("module", "events").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	topics := GetTopics(conf)

	return &EventFeed{
		log:       logger,
		Config:    conf,
		EthClient: ethClient,
		Topics:    topics,
		Wal:       wal,
		ds:        ds,
	}, nil
}

func (e *EventFeed) Listen(
	ctx context.Context,
) error {
	e.log.Debug().Msg("starting event feed")

	headers, err := e.listenForBlockHeaders(ctx)
	if err != nil {
		return err
	}
	e.ProcessBlocks(ctx, headers)

	return nil
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
