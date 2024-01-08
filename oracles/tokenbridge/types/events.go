package types

import (
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/serialize"
)

type TokenBridgeEventType int64

const (
	DEPOSITS TokenBridgeEventType = iota
)

func (et TokenBridgeEventType) String() string {
	switch et {
	case DEPOSITS:
		return "Deposit"
	default:
		return "Unknown"
	}
}

func (et TokenBridgeEventType) Uint64() uint64 {
	return uint64(et)
}

type DepositEvent struct {
	Sender    string   `json:"sender"`
	Amount    *big.Int `json:"amount"`
	TxHash    string   `json:"txHash"`
	BlockHash string   `json:"blockHash"`
	ChainID   string   `json:"chainId"`
	// probably don't need BlockHash and ChainID
}

func (e *DepositEvent) MarshalBinary() ([]byte, error) {
	return serialize.Encode(e)
}

func (e *DepositEvent) UnmarshalBinary(data []byte) error {
	return serialize.DecodeInto(data, e)
}
