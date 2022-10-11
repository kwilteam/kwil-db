package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"kwil/pkg/types/chain/pricing"
	v0 "kwil/x/api/v0"
	"kwil/x/logx"
)

type Service struct {
	v0.UnimplementedKwilServiceServer

	ds      DepositStore
	log     logx.Logger
	pricing pricing.PriceBuilder
	cc      ContractClient
}

type DepositStore interface {
	GetBalance(address string) (*big.Int, error)
	SetBalance(address string, balance *big.Int) error
	GetSpent(addr string) (*big.Int, error)
	SetSpent(addr string, amt *big.Int) error
}

type ContractClient interface {
	ReturnFunds(ctx context.Context, recip common.Address, amt *big.Int, fee *big.Int) (*types.Transaction, error)
}

func NewService(ds DepositStore, p pricing.PriceBuilder, cc ContractClient) *Service {
	return &Service{
		ds:      ds,
		pricing: p,
		log:     logx.New(),
		cc:      cc,
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
		s.log.Debug("fee is too low for the requested operation")
		return nil, ErrFeeTooLow
	}

	// get the balance of the sender
	bal, err := s.ds.GetBalance(*from)
	if err != nil {
		return fb, fmt.Errorf("failed to get balance for %s: %w", *from, err)
	}

	// check if the balance is greater than the fee
	if fee.Cmp(bal) > 0 {
		return fb, ErrNotEnoughFunds
	}

	// TODO: Write to WAL
	// I figured this isn't critically important for a test version where they use fake funds

	fb = bal.Sub(bal, cost)

	return fb, nil
}
