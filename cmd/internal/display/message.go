package display

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// RespTxHash is used to represent a transaction hash in cli
// NOTE: it's different from transactions.TxHash, this is for display purpose.
type RespTxHash []byte

func (h RespTxHash) Hex() string {
	return hex.EncodeToString(h)
}

func (h RespTxHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TxHash string `json:"tx_hash"`
	}{
		TxHash: h.Hex(),
	})
}

func (h RespTxHash) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("TxHash: %s", h.Hex())), nil
}
