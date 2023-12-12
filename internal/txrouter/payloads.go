package txrouter

import (
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/serialize"
)

// payloads includes extra payloads that are not included in core/types/transactions

type CreditPayload struct {
	DepositID string   // the transaction ID from the source chain that created the deposit
	Account   []byte   // the account to be credited
	Amount    *big.Int // the amount of tokens to be credited
}

func (p *CreditPayload) MarshalBinary() ([]byte, error) {
	return serialize.Encode(p)
}

func (p *CreditPayload) UnmarshalBinary(data []byte) error {
	return serialize.DecodeInto(data, p)
}
