package txrouter

import "github.com/kwilteam/kwil-db/core/types/serialize"

// payloads includes extra payloads that are not included in core/types/transactions

type CreditPayload struct{}

func (p *CreditPayload) MarshalBinary() ([]byte, error) {
	return serialize.Encode(p)
}

func (p *CreditPayload) UnmarshalBinary(data []byte) error {
	return serialize.DecodeInto(data, p)
}
