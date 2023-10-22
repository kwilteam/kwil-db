package abci

import (
	"errors"

	"github.com/kwilteam/kwil-db/core/types/transactions"
)

var (
	// ErrTxNotFound is indicates when the a transaction was not found in the
	// nodes blocks or mempool.
	ErrTxNotFound = errors.New("transaction not found")
)

type TxCode = transactions.TxCode

// consumers should use the transactions package codes
const (
	CodeOk               = transactions.CodeOk
	CodeEncodingError    = transactions.CodeEncodingError
	CodeInvalidTxType    = transactions.CodeInvalidTxType
	CodeInvalidSignature = transactions.CodeInvalidSignature
	CodeInvalidNonce     = transactions.CodeInvalidNonce
	CodeWrongChain       = transactions.CodeWrongChain
	CodeUnknownError     = transactions.CodeUnknownError
)
