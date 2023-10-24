package transactions

import "errors"

// TransactionResult is the result of a transaction execution on chain
type TransactionResult struct {
	Code      uint32   `json:"code"`
	Log       string   `json:"log"`
	GasUsed   int64    `json:"gas_used"`
	GasWanted int64    `json:"gas_wanted"`
	Data      []byte   `json:"data,omitempty"`
	Events    [][]byte `json:"events,omitempty"`
}

var (
	// ErrTxNotFound is indicates when the a transaction was not found in the
	// nodes blocks or mempool.
	ErrTxNotFound   = errors.New("transaction not found")
	ErrWrongChain   = errors.New("wrong chain ID")
	ErrInvalidNonce = errors.New("invalid nonce")
)

type TxCode uint32

const (
	CodeOk               TxCode = 0
	CodeEncodingError    TxCode = 1
	CodeInvalidTxType    TxCode = 2
	CodeInvalidSignature TxCode = 3
	CodeInvalidNonce     TxCode = 4
	CodeWrongChain       TxCode = 5
	CodeUnknownError     TxCode = 6
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
	case CodeInvalidNonce:
		return "invalid nonce"
	case CodeWrongChain:
		return "wrong chain"
	default:
		return "unknown tx error"
	}
}

// TcTxQueryResponse is the response of a transaction query
// NOTE: not `txpb.TxQueryResponse` so TransportClient only use our brewed type
// same as `TransactionResult`
type TcTxQueryResponse struct {
	Hash     []byte            `json:"hash,omitempty"`
	Height   int64             `json:"height,omitempty"`
	Tx       Transaction       `json:"tx"`
	TxResult TransactionResult `json:"tx_result"`
}
