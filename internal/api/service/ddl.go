package service

import (
	"context"
	"errors"
	apitypes "github.com/kwilteam/kwil-db/internal/api/types"
	"github.com/kwilteam/kwil-db/internal/crypto"
)

func (s *Service) AlterDatabase(ctx context.Context, m *apitypes.AlterDatabaseMsg) error {
	/*
		Service function for altering a database

		Currently supporting:
		 - Creating tables
		 - Altering tables
		 - Dropping tables

		 - Creating Parameterized Queries
		 - Altering Parameterized Queries
		 - Dropping Parameterized Queries

		First, we check the signature of the message.

		Then, we get the cost from the config.
		If it does not exist, we return an error.

		Then, we check the balances.
		If valid, we set the new balance and forward the message to cosmos.
	*/

	// Check ID
	if !m.CheckID() {
		return apitypes.ErrInvalidID
	}

	// Check the signature
	valid, err := crypto.CheckSignature(m.From, m.Signature, []byte(m.ID))
	if err != nil {
		return err
	}
	if !valid {
		return apitypes.ErrInvalidSignature
	}

	// Validate the balances
	amt, err := s.validateBalances(&m.From, &m.Operation, &m.Crud, &m.Fee)
	if err != nil {
		return err
	}

	err = s.ds.SetBalance(m.From, amt)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to set balance for %s", m.From)
		return err
	}

	// Forward message to Kafka

	return nil
}

var ErrInvalidDDLType = errors.New("invalid ddl type")
