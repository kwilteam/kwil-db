package abci

import "errors"

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
