package transactions

import "errors"

// TransactionResult is the result of a transaction execution on chain
type TransactionResult struct {
	Code      uint32   `json:"code,omitempty"`
	Log       string   `json:"log,omitempty"`
	GasUsed   int64    `json:"gas_used,omitempty"`
	GasWanted int64    `json:"gas_wanted,omitempty"`
	Data      []byte   `json:"data,omitempty"`
	Events    [][]byte `json:"events,omitempty"`
}

var (
	// ErrTxNotFound is indicates when the a transaction was not found in the
	// nodes blocks or mempool.
	ErrTxNotFound = errors.New("transaction not found")
)

type TxCode uint32

const (
	CodeOk               TxCode = 0
	CodeEncodingError    TxCode = 1
	CodeInvalidTxType    TxCode = 2
	CodeInvalidSignature TxCode = 3
	CodeInvalidNonce     TxCode = 4
	CodeUnknownError     TxCode = 5 // for now it's for all non-encoding error
)

func (c TxCode) Uint32() uint32 {
	return uint32(c)
}

func (tc TxCode) String() string {
	switch tc {
	case CodeOk:
		return "OK"
	case CodeEncodingError:
		return "encoding error"
	case CodeInvalidTxType:
		return "invalid tx type"
	case CodeInvalidSignature:
		return "invalid signature"
	default:
		return "unknown tx error"
	}
}

type TcTxQueryResponse struct {
	Hash     []byte
	Height   int64
	Tx       Transaction
	TxResult TransactionResult
}
