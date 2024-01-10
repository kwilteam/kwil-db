package ethbridge

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/internal/voting"
)

type AccountCredit struct {
	Account   string
	Amount    *big.Int
	TxHash    string
	BlockHash string
	ChainID   string
}

func (ac *AccountCredit) MarshalBinary() ([]byte, error) {
	return serialize.Encode(ac)
}

func (ac *AccountCredit) UnmarshalBinary(data []byte) error {
	return serialize.DecodeInto(data, ac)
}

func (ac *AccountCredit) Type() string {
	return "AccountCredit"
}

func (ac *AccountCredit) Apply(ctx context.Context, datastores *voting.Datastores) error {
	err := datastores.Accounts.Credit(ctx, []byte(ac.Account), ac.Amount)
	if err != nil {
		return err
	}
	fmt.Println("Credited account for deposit, account: ", ac.Account, " amount: ", ac.Amount.String())

	return nil
}
