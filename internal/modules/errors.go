package modules

import (
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// ABCIModuleError is a way for an ABCI module to optionally provide execution
// detail back to the ABCI application. Use with ErrCode or ConvertModuleError.
type ABCIModuleError struct {
	Code   transactions.TxCode
	Detail string
	// GasUsed *big.Int // failed execution can use gas, consider this approach
}

func (me ABCIModuleError) Error() string {
	return fmt.Sprintf("code %d: %s", me.Code, me.Detail)
}

// ErrCode gets the transaction code if the error is an ABCIModuleError. If the
// error is nil, it returns CodeOK. If the error is not an ABCIModuleError, it
// returns CodeUnknownError. Otherwise, it returns the code in the typed error.
func ErrCode(err error) transactions.TxCode {
	if err == nil {
		return transactions.CodeOk
	}
	me := ConvertModuleError(err)
	if me == nil {
		return transactions.CodeUnknownError
	}
	return me.Code
}

// ConvertModuleError extracts a ABCIModuleError from the error. If it is not a
// ABCIModuleError, it returns nil.
func ConvertModuleError(err error) *ABCIModuleError {
	var modErr ABCIModuleError
	if errors.As(err, &modErr) {
		return &modErr
	}
	var modErrP *ABCIModuleError
	if errors.As(err, &modErrP) {
		return modErrP
	}
	return nil
}
