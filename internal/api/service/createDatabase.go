package service

import (
	"context"
	"errors"
	"github.com/kwilteam/kwil-db/pkg/types"
	"math/big"
)

// CreateDatabase Service Function for CreateDatabase
func (s *Service) CreateDatabase(ctx context.Context, db *types.CreateDatabase) error {

	/*
		Service Function for CreateDatabase

		First, we need to check the cost associated with creating a database.
		Right now, this will be a static cost set in the config file. config.cost.database.create.

		If the user has enough funds, we will subtract the fund and propagate to Cosmos to reach consensus.
		(this is where a WAL should be used and rollback the funds if a crash occurs)

		The request can have an optional flag to wait to send a response until the database is created,
		or to respond once it has been written to the Wal and sent to cosmos.

	*/

	// First, check the signature
	// TODO: check the signature

	// Convert cost from string to big.Int
	cost := new(big.Int)
	cost, ok := cost.SetString(s.conf.Cost.Database.Create, 10)
	if !ok {
		return errors.New("failed to parse cost")
	}

	// convert the sent fee from string to big.Int
	fee := new(big.Int)
	fee, ok = fee.SetString(db.Fee, 10)
	if !ok {
		return errors.New("failed to parse fee")
	}

	// compare the cost to what is sent
	if cost.Cmp(fee) > 0 {
		s.log.Debug().Msg("fee is too low for the requested operation")
		return ErrFeeTooLow
	}

	// Get the balance of the sender
	bal, err := s.Ds.GetBalance(db.From)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to get balance for %s", db.From)
		return err // it is ok to return this error since the handler never returns errors to the client
	}

	// Check if the balance is greater than the fee
	if fee.Cmp(bal) > 0 {
		s.log.Debug().Msg("not enough funds")
		return ErrNotEnoughFunds
	}

	// TODO: Write to WAL

	retBal := bal.Sub(bal, cost)
	err = s.Ds.SetBalance(db.From, retBal)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to set balance for %s", db.From)
		return err
	}

	// TODO: Send to Cosmos

	return nil
}
