package display

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type TxHashResponse struct {
	TxHash string `json:"tx_hash"`
}

// RespTxHash is used to represent a transaction hash in cli
// NOTE: it's different from transactions.TxHash, this is for display purpose.
// It implements the MsgFormatter interface
type RespTxHash []byte

func (h RespTxHash) Hex() string {
	return hex.EncodeToString(h)
}

func (h RespTxHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(TxHashResponse{TxHash: h.Hex()})
}

func (h RespTxHash) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("TxHash: %s", h.Hex())), nil
}

// RespString is used to represent a string in cli
// It implements the MsgFormatter interface
type RespString string

func (s RespString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s)) // must convert to string to avoid infinite recursion
}

func (s RespString) MarshalText() ([]byte, error) {
	return []byte(s), nil
}
