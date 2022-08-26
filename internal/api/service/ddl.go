package service

import (
	"context"
	"errors"
	"github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/kwilteam/kwil-db/pkg/types"
)

func (s *Service) DDL(ctx context.Context, ddl *types.DDL) error {

	/*
		Service function for DDL

		First, we check the signature of the message.

		Then, we get the cost from the config.
		If it does not exist, we return an error.

		Then, we check the balances.
		If valid, we set the new balance and forward the message to cosmos.
	*/

	// Check the signature
	valid, err := crypto.CheckSignature(ddl.From, ddl.Signature, []byte(ddl.Id))
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidSignature
	}

	var cost string
	// Find the type of DDL they are submitting
	switch ddl.Type {
	case "table_create":
		cost = s.conf.Cost.Ddl.Table.Create
	case "table_delete":
		cost = s.conf.Cost.Ddl.Table.Delete
	case "table_modify":
		cost = s.conf.Cost.Ddl.Table.Modify
	case "role_create":
		cost = s.conf.Cost.Ddl.Role.Create
	case "role_delete":
		cost = s.conf.Cost.Ddl.Role.Delete
	case "role_modify":
		cost = s.conf.Cost.Ddl.Role.Modify
	case "query_create":
		cost = s.conf.Cost.Ddl.Query.Create
	case "query_delete":
		cost = s.conf.Cost.Ddl.Query.Delete

	default:
		return ErrInvalidDDLType
	}

	// Check the balance
	amt, err := s.validateBalances(&ddl.From, &cost, &ddl.Fee)
	if err != nil {
		return err
	}

	// Set the new balance
	err = s.ds.SetBalance(ddl.From, amt)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to set balance for %s", ddl.From)
		return err
	}

	// TODO: Send to Cosmos

	return nil
}

var ErrInvalidDDLType = errors.New("invalid ddl type")
