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

// These aliases are for internal use by the ABCI application. They exist only
// for brevity in the application code.
const (
	codeOk                  = transactions.CodeOk
	codeEncodingError       = transactions.CodeEncodingError
	codeInvalidTxType       = transactions.CodeInvalidTxType
	codeInvalidSignature    = transactions.CodeInvalidSignature
	codeInvalidNonce        = transactions.CodeInvalidNonce
	codeWrongChain          = transactions.CodeWrongChain
	codeInsufficientBalance = transactions.CodeInsufficientBalance
	codeInsufficientFee     = transactions.CodeInsufficientFee
	codeInvalidAmount       = transactions.CodeInvalidAmount
	codeUnknownError        = transactions.CodeUnknownError
)
