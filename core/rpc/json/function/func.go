package function

import (
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/core/types"
)

// Methods

const (
	MethodVersion   jsonrpc.Method = "function.version"
	MethodVerifySig jsonrpc.Method = "function.verify_sig"
)

// Requests

type VerifySignatureRequest struct {
	Signature *TxSignature   `json:"signature,omitempty"`
	Sender    types.HexBytes `json:"sender,omitempty"`
	Msg       types.HexBytes `json:"msg,omitempty"`
}

type TxSignature struct {
	SignatureBytes []byte `json:"sig,omitempty"`
	SignatureType  string `json:"type,omitempty"`
}

// Responses

type VerifySignatureResponse struct {
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}
