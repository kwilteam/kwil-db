package service

import (
	"errors"
	"math/big"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/kwilteam/kwil-db/pkg/types/chain/pricing"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Service struct {
	v0.UnimplementedKwilServiceServer

	ds      DepositStore
	log     zerolog.Logger
	pricing pricing.PriceBuilder
}

type DepositStore interface {
	GetBalance(address string) (*big.Int, error)
	SetBalance(address string, balance *big.Int) error
}

func NewService(ds DepositStore, p pricing.PriceBuilder) *Service {
	logger := log.With().Str("module", "service").Logger()
	return &Service{
		log:     logger,
		ds:      ds,
		pricing: p,
	}
}

// validateBalances checks to ensure that the sender has enough funds to cover the fee.
// It also checks to ensure that the fee is not too low.
// Finally, it returns what the new balance should be if the operation is to be executed.
// It also returns an error if the amount is not enough
func (s *Service) validateBalances(from *string, op *int32, cr *int32, fe *string) (*big.Int, error) {

	fb := big.NewInt(0) // final balance

	// get the cost of the operation
	c := s.pricing.Operation(byte(*op)).Crud(byte(*cr)).Build()

	// convert cost from int64 to big.Int
	cost := big.NewInt(c)

	// convert the sent fee from string to big.Int
	fee := new(big.Int)
	fee, ok := fee.SetString(*fe, 10)
	if !ok {
		return fb, errors.New("failed to parse fee")
	}

	// compare the cost to what is sent
	if cost.Cmp(fee) > 0 {
		s.log.Debug().Msg("fee is too low for the requested operation")
		return fb, ErrFeeTooLow
	}

	// get the balance of the sender
	bal, err := s.ds.GetBalance(*from)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to get balance for %s", *from)
		return fb, err // it is ok to return this error since the handler never returns errors to the client
	}

	// check if the balance is greater than the fee
	if fee.Cmp(bal) > 0 {
		s.log.Debug().Msg("not enough funds")
		return fb, ErrNotEnoughFunds
	}

	// TODO: Write to WAL
	// I figured this isn't critically important for a test version where they use fake funds

	fb = bal.Sub(bal, cost)

	return fb, nil
}
