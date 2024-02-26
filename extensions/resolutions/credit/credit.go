// package credit implements a credit resolution, allowing accounts to be credited with a given amount.
package credit

import (
	"context"
	"errors"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/functions"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
)

const CreditAccountEventType = "credit_account"

// use golang's init function, which runs before main, to register the extension
// see more here: https://www.digitalocean.com/community/tutorials/understanding-init-in-go
func init() {
	// calling RegisterResolution will make the resolution available on startup
	err := resolutions.RegisterResolution(CreditAccountEventType, resolutions.ResolutionConfig{
		// Setting the refund resolution to 1/3 will refund all validators who spent gas voting on
		// the resolution if the resolution does not pass, but receives at least 1/3 of the total
		// network voting power. This is useful for accounting for soft forks on chains like Ethereum;
		// if 1/3 of the Kwil validators vote on a resolution due to data in a soft fork, they can
		// still get refunded. This does not cover all edge cases, but helps to minimize the risk of
		// validators losing money when acting in good faith.
		RefundThreshold: big.NewRat(1, 3),
		// Setting the confirmation threshold to 2/3 will require 2/3 of the total network voting
		// power to vote on the resolution in order for it to pass. If 2/3 of the network votes on
		// a resolution of this type, then it will be applied. If less than 2/3 of the network votes
		// have voted on this resolution by expiration, the resolution will fail.
		ConfirmationThreshold: big.NewRat(2, 3),
		// Setting the expiration height to 600 gives the validators approximately 1 hour to vote
		// on the resolution after it has been created (assuming 6s block times). If the resolution
		// has not received enough votes by the expiration height, it will fail.
		ExpirationPeriod: 600,
		// ResolveFunc defines what will happen if the resolution is approved by the network.
		// For the credit account oracle, we will credit the account with the given amount.
		// The amount cannot be negative.
		ResolveFunc: func(ctx context.Context, app *common.App, resolution *resolutions.Resolution) error {
			// Unmarshal the resolution payload into an AccountCreditResolution
			var credit AccountCreditResolution
			err := credit.UnmarshalBinary(resolution.Body)
			if err != nil {
				return err
			}

			if credit.Amount.Sign() < 0 {
				return errors.New("credit amount cannot be negative")
			}

			// Credit the account with the given amount
			return functions.Accounts.Credit(ctx, app.DB, credit.Account, credit.Amount)
		},
	})
	if err != nil {
		panic(err)
	}
}

// AccountCreditResolution is a resolution that allows accounts to be credited with a given amount.
// It is used by both the credit_account resolution and the eth_deposit_oracle.
// It can be serialized and deserialized to be passed around the network.
// The amount cannot be negative, as this will fail RLP encoding.
type AccountCreditResolution struct {
	// Account is the account to be credited.
	// This can be an Ethereum address (decoded from hex), a validator ed25519 public key,
	// or any other custom account identifier implemented in an auth extension.
	Account []byte
	// Amount is the amount to be credited to the account.
	// It uses a big.Int to allow for arbitrary precision, and to allow for uint256 values,
	// which are commonly used in token contracts on Ethereum.
	Amount *big.Int
	// TxHash is the hash of the Ethereum transaction that emitted the EVM event to credit the account.
	// This ensures that, even if the same account is credited the same amount multiple times,
	// that each credit resolution is unique. It is critical that all resolutions in Kwil are
	// unique, as they are idempotent for the lifetime of the entire network.
	TxHash []byte
}

// MarshalBinary marshals the AccountCreditResolution to binary.
// We do not use the popular json.Marshal library because we need this serialization
// to be deterministic. Kwil contains a serialization library that uses Ethereum's
// RLP encoding, which is deterministic and used for all serialization in Kwil.
func (a *AccountCreditResolution) MarshalBinary() ([]byte, error) {
	return serialize.Encode(a)
}

// UnmarshalBinary unmarshals the AccountCreditResolution from binary.
// It is the inverse of MarshalBinary, and uses the same serialization library.
func (a *AccountCreditResolution) UnmarshalBinary(data []byte) error {
	return serialize.DecodeInto(data, a)
}
