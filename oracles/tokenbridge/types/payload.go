package types

import (
	"context"
	"encoding"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

type AccountCredit struct {
	Account   string // as they would set in tx.Sender
	Amount    *big.Int
	TxHash    string
	BlockHash string
	ChainID   string
}

var _ encoding.BinaryUnmarshaler = (*AccountCredit)(nil)
var _ encoding.BinaryMarshaler = (*AccountCredit)(nil)

func (ac *AccountCredit) UnmarshalBinary(b []byte) error {
	return serialize.DecodeInto(b, ac)
}

func (ac *AccountCredit) MarshalBinary() ([]byte, error) {
	return serialize.Encode(ac)
}

func (ac *AccountCredit) Apply(ctx context.Context, datastores *types.Datastores) error {
	err := datastores.Accounts.Credit(ctx, []byte(ac.Account), ac.Amount)
	if err != nil {
		return err
	}
	fmt.Println("Credited account for deposit, account: ", ac.Account, " amount: ", ac.Amount.String())

	// Try accessing the database to see if it's working
	// if datastores.Databases != nil {
	// 	results, err := datastores.Databases.Execute(ctx, "x815b03a67203ce32fd4d0ef5e816d501e5023ad0c1cc7f926a45fddc", "SELECT * FROM users", nil)
	// 	if err != nil {
	// 		fmt.Println("Error accessing database: ", err)
	// 	}

	// 	fmt.Println("Database results: ", results.Columns(), "    ", results.Rows())
	// }

	return nil
}

func (ac *AccountCredit) Type() string {
	return "AccountCredit"
}
