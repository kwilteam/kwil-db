package service

import (
	"context"
	"errors"
<<<<<<< HEAD
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
=======
	"github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/kwilteam/kwil-db/pkg/types"
)

func (s *Service) DDL(ctx context.Context, ddl *types.DDL) error {

	/*
		Service function for DDL
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

		First, we check the signature of the message.

		Then, we get the cost from the config.
		If it does not exist, we return an error.

		Then, we check the balances.
		If valid, we set the new balance and forward the message to cosmos.
	*/

<<<<<<< HEAD
	// Check ID
	if !m.CheckID() {
		return apitypes.ErrInvalidID
	}

	// Check the signature
	valid, err := crypto.CheckSignature(m.From, m.Signature, []byte(m.ID))
=======
	// Check the signature
	valid, err := crypto.CheckSignature(ddl.From, ddl.Signature, []byte(ddl.Id))
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		return err
	}
	if !valid {
<<<<<<< HEAD
		return apitypes.ErrInvalidSignature
	}

	// Validate the balances
	amt, err := s.validateBalances(&m.From, &m.Operation, &m.Crud, &m.Fee)
=======
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
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		return err
	}

<<<<<<< HEAD
	err = s.ds.SetBalance(m.From, amt)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to set balance for %s", m.From)
		return err
	}

	// Forward message to Kafka
=======
	// Set the new balance
	err = s.ds.SetBalance(ddl.From, amt)
	if err != nil {
		s.log.Debug().Err(err).Msgf("failed to set balance for %s", ddl.From)
		return err
	}

	// TODO: Send to Cosmos
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

	return nil
}

var ErrInvalidDDLType = errors.New("invalid ddl type")
