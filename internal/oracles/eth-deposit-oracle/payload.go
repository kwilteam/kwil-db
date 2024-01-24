package deposit_oracle

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/internal/voting"
	"go.uber.org/zap"
)

const (
	// AccountCreditResolutionReward is the amount credited to the voter for participating in the voting process for AccountCredit Resolution.
	// This acts as incentive for voters to run required resources and participate in the voting process for AccountCredit Resolution
	// This gets credited to the voter's account only when the AccountCredit Resolution is approved by supermajority of the validators.
	AccountCreditResolutionReward = 1000
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

func (ac *AccountCredit) Apply(ctx context.Context, datastores voting.Datastores, proposer []byte, voters []voting.Voter, logger log.Logger) error {
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
	logger.Debug("Credited account for deposit", zap.String("account", ac.Account), zap.String("amount", ac.Amount.String()))

	for _, voter := range voters {
		// credit the voter
		err = datastores.Accounts.Credit(ctx, voter.PubKey, big.NewInt(AccountCreditResolutionReward))
		if err != nil {
			return err
		}
		logger.Debug("Rewarded voter for voting for AccountCredit event", zap.String("voter", hex.EncodeToString(voter.PubKey)))
	}

	return nil
}
