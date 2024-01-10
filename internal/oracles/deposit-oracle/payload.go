package deposit_oracle

import (
	"context"
	"encoding/hex"
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
	// trim the 0x prefix
	if len(ac.Account) > 2 && ac.Account[:2] == "0x" {
		ac.Account = ac.Account[2:]
	} else {
		return fmt.Errorf("account address must start with 0x")
	}

	// decode the hex string into a byte slice
	bts, err := hex.DecodeString(ac.Account)
	if err != nil {
		return err
	}

	err = datastores.Accounts.Credit(ctx, bts, ac.Amount)
	if err != nil {
		return err
	}
	fmt.Println("Credited account for deposit, account: ", ac.Account, " amount: ", ac.Amount.String())

	return nil
}
