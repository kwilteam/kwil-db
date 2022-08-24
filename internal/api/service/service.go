package service

import (
	"errors"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/big"
)

// Struct for service logic
type Service struct {
	conf *types.Config
	Ds   DepositStore
	log  zerolog.Logger
}

type DepositStore interface {
	GetBalance(address string) (*big.Int, error)
	SetBalance(address string, balance *big.Int) error
}

// NewService returns a pointer Service.
func NewService(conf *types.Config, ds DepositStore) *Service {
	logger := log.With().Str("module", "service").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	return &Service{
		log:  logger,
		conf: conf,
		Ds:   ds,
	}
}

var ErrNotEnoughFunds = errors.New("not enough funds")
var ErrFeeTooLow = errors.New("the sent fee is too low for the requested operation")
