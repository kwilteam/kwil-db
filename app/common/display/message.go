package display

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// TxHashAndExecResponse is meant to combine the "tx_hash" marshalling of
// RespTxHash with a RespTxQuery in an "exec_result" field.
type TxHashAndExecResponse struct {
	Hash      RespTxHash   // embedding breaks MarshalJSON of composing types
	QueryResp *RespTxQuery `json:"exec_result"`
}

// NewTxHashAndExecResponse makes a TxHashAndExecResponse from a TcTxQueryResponse.
func NewTxHashAndExecResponse(resp *types.TcTxQueryResponse) *TxHashAndExecResponse {
	return &TxHashAndExecResponse{
		Hash:      RespTxHash(resp.Hash),
		QueryResp: &RespTxQuery{Msg: resp},
	}
}

func (h *TxHashAndExecResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TxHash    string       `json:"tx_hash"`
		QueryResp *RespTxQuery `json:"exec_result"`
	}{
		TxHash:    h.Hash.Hex(),
		QueryResp: h.QueryResp,
	})
}

// MarshalText deduplicates the tx hash for a compact readable output, unlike
// the JSON marshalling that is meant to be a composition of both RespTxHash and
// RespTxQuery.
func (h TxHashAndExecResponse) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`TxHash: %s
Status: %s
Height: %d
Log: %s`, h.Hash.Hex(),
		heightStatus(h.QueryResp.Msg),
		h.QueryResp.Msg.Height,
		h.QueryResp.Msg.TxResult.Log,
	),
	), nil
}

var _ MsgFormatter = (*TxHashAndExecResponse)(nil)
var _ MsgFormatter = (*RespTxQuery)(nil)

type TxHashResponse struct {
	TxHash string `json:"tx_hash"`
}

// RespTxHash is used to represent a transaction hash in cli
// NOTE: it's different from types.TxHash, this is for display purpose.
// It implements the MsgFormatter interface
type RespTxHash types.Hash

func (h RespTxHash) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h RespTxHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(TxHashResponse{TxHash: h.Hex()})
}

func (h RespTxHash) MarshalText() ([]byte, error) {
	return []byte("TxHash: " + h.Hex()), nil
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

// RespTxQuery is used to represent a transaction response in cli
type RespTxQuery struct {
	Msg     *types.TcTxQueryResponse
	WithRaw bool
}

func (r *RespTxQuery) MarshalJSON() ([]byte, error) {
	out := &struct {
		Hash     string                  `json:"hash"` // HEX
		Height   int64                   `json:"height"`
		Tx       *types.Transaction      `json:"tx"`
		TxResult types.TransactionResult `json:"tx_result"`
		Raw      string                  `json:"raw,omitempty"`
		Warn     string                  `json:"warning,omitempty"`
	}{
		Hash:     r.Msg.Hash.String(),
		Height:   r.Msg.Height,
		Tx:       r.Msg.Tx,
		TxResult: r.Msg.TxResult,
	}
	// Always try to serialize to verify hash, but only show raw if requested.
	if r.Msg.Tx != nil {
		raw, err := r.Msg.Tx.MarshalBinary()
		if err != nil {
			out.Warn = "ERROR encoding transaction: " + err.Error()
		} else if r.WithRaw {
			out.Raw = hex.EncodeToString(raw)
			hash := sha256.Sum256(raw)
			if hash != r.Msg.Hash {
				out.Warn = fmt.Sprintf("HASH MISMATCH: requested %s; received %x",
					r.Msg.Hash, hash)
			}
		}
	}
	return json.Marshal(out)
}

func heightStatus(res *types.TcTxQueryResponse) string {
	status := "failed"
	if res.Height == -1 {
		status = "pending"
	} else if res.TxResult.Code == uint32(types.CodeOk) {
		status = "success"
	}
	return status
}

func (r *RespTxQuery) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf(`Transaction ID: %s
Status: %s
Height: %d
Log: %s`,
		r.Msg.Hash.String(),
		heightStatus(r.Msg),
		r.Msg.Height,
		r.Msg.TxResult.Log,
	)

	// Always try to serialize to verify hash, but only show raw if requested.
	if r.Msg.Tx == nil {
		return []byte(msg), nil
	}

	raw, err := r.Msg.Tx.MarshalBinary()
	if err != nil {
		msg += "\nERROR encoding transaction: " + err.Error()
	} else {
		if r.WithRaw {
			msg += "\nRaw: " + hex.EncodeToString(raw)
		}
		hash := sha256.Sum256(raw)
		if hash != r.Msg.Hash {
			msg += fmt.Sprintf("\nWARNING! HASH MISMATCH:\n\tRequested %s\n\tReceived  %x",
				r.Msg.Hash, hash)
		}
	}

	return []byte(msg), nil
}
