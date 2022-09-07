package service

<<<<<<< HEAD
/*
	I broke createDatabase into it's own service function in case there is special logic that needs to occur
	An example would be that, in the future, a database might only need to come to consensus within itself (as opposed to on the global network)
*/

import (
	"context"
	apitypes "github.com/kwilteam/kwil-db/internal/api/types"
	"github.com/kwilteam/kwil-db/internal/crypto"
	"strings"
)

// CreateDatabase Service Function for CreateDatabase
func (s *Service) CreateDatabase(ctx context.Context, db *apitypes.CreateDatabaseMsg) error {
=======
import (
	"context"
	"github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/kwilteam/kwil-db/pkg/types"
)

// CreateDatabase Service Function for CreateDatabase
func (s *Service) CreateDatabase(ctx context.Context, db *types.CreateDatabase) error {
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

	/*
		Service Function for CreateDatabase

<<<<<<< HEAD
		First, we check the operation and crud are both 0 (see the the pricing package in prices.go for info on convertion operations to bytes)

		Next, we reconstruct and check the id

		Next, we check the incoming signature.
=======
		First, we check the incoming signature.
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
		If valid, we then validate the balances (validates the sent fee, cost, and sender balance)

		If valid, we set the new balance and forward the message to cosmos

	*/

<<<<<<< HEAD
	// check that operation and crud are valid
	if int8(db.Operation) != 0 {
		return apitypes.ErrIncorrectOperation
	}

	if int8(db.Crud) != 0 {
		return apitypes.ErrIncorrectCrud
	}

	// check ID
	if !db.CheckID() {
		return apitypes.ErrInvalidID
	}

	//  check the signature
	valid, err := crypto.CheckSignature(db.From, db.Signature, []byte(db.ID))
=======
	// First, check the signature
	valid, err := crypto.CheckSignature(db.From, db.Signature, []byte(db.Id))
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		return err
	}
	if !valid {
<<<<<<< HEAD
		return apitypes.ErrInvalidSignature
	}

	// Next, check the balances
	amt, err := s.validateBalances(&db.From, &db.Operation, &db.Crud, &db.Fee)
=======
		return ErrInvalidSignature
	}

	// Next, check the balances
	amt, err := s.validateBalances(&db.From, &s.conf.Cost.Database.Create, &db.Fee)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
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
<<<<<<< HEAD

func createDBID(owner, name, fee, dbtype string) []byte {
	sb := strings.Builder{}
	sb.WriteString(owner)
	sb.WriteString(name)
	sb.WriteString(fee)
	sb.WriteString(dbtype)

	return crypto.Sha384([]byte(sb.String()))
}
=======
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
