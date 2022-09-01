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

		First, we check the incoming signature.
		If valid, we then validate the balances (validates the sent fee, cost, and sender balance)

		If valid, we set the new balance and forward the message to cosmos

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

	return nil
}
