package service

/*
	I broke createDatabase into it's own service function in case there is special logic that needs to occur
	An example would be that, in the future, a database might only need to come to consensus within itself (as opposed to on the global network)
*/

import (
	"context"
	"strings"

	types "github.com/kwilteam/kwil-db/internal/api/types"
	"github.com/kwilteam/kwil-db/internal/chain/crypto"
)

// CreateDatabase Service Function for CreateDatabase
func (s *Service) CreateDatabase(ctx context.Context, db *types.CreateDatabaseMsg) error {

	/*
		Service Function for CreateDatabase

		First, we check the operation and crud are both 0 (see the the pricing package in prices.go for info on convertion operations to bytes)

		Next, we reconstruct and check the id

		Next, we check the incoming signature.
		If valid, we then validate the balances (validates the sent fee, cost, and sender balance)

		If valid, we set the new balance and forward the message to cosmos

	*/

	// check that operation and crud are valid
	if int8(db.Operation) != 0 {
		return types.ErrIncorrectOperation
	}

	if int8(db.Crud) != 0 {
		return types.ErrIncorrectCrud
	}

	// check ID
	if !db.CheckID() {
		return types.ErrInvalidID
	}

	//  check the signature
	valid, err := crypto.CheckSignature(db.From, db.Signature, []byte(db.ID))
	if err != nil {
		return err
	}
	if !valid {
		return types.ErrInvalidSignature
	}

	// Next, check the balances
	amt, err := s.validateBalances(&db.From, &db.Operation, &db.Crud, &db.Fee)
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

// this is used in the package tests
func createDBID(owner, name, fee string) []byte {
	sb := strings.Builder{}
	sb.WriteString(owner)
	sb.WriteString(name)
	sb.WriteString(fee)

	return crypto.Sha384([]byte(sb.String()))
}
