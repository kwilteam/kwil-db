package service

import (
	"context"
	"github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/kwilteam/kwil-db/pkg/types"
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
	valid, err := crypto.CheckSignature(db.From, db.Signature, []byte(db.Id))
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidSignature
	}

	// Next, check the balances
	amt, err := s.validateBalances(&db.From, &s.conf.Cost.Database.Create, &db.Fee)
	if err != nil {
		return err
	}

	err = s.ds.SetBalance(db.From, amt)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to set balance for %s", db.From)
		return err
	}

	// TODO: Send to Cosmos

	return nil
}
