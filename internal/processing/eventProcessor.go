package processing

import (
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/big"
)

type DepositStore interface {
	Deposit(amount *big.Int, addr string) error
	GetBalance(addr string) (*big.Int, error)
}

type EventProcessor struct {
	log       zerolog.Logger
	EventChan chan map[string]interface{}
	Conf      *types.Config
	Deposits  DepositStore
}

func New(conf *types.Config, ech chan map[string]interface{}, ds DepositStore) *EventProcessor {
	logger := log.With().Str("module", "processing").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	return &EventProcessor{
		log:       logger,
		EventChan: ech,
		Conf:      conf,
		Deposits:  ds,
	}
}
