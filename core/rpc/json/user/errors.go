package userjson

import "fmt"

// BroadcastError is a structured error object used by MethodBroadcast when
// creating a Response with this in Error.Data. This error type would typically
// be in an response with the ErrorTxExecFailure RPC ErrorCode.
type BroadcastError struct {
	ErrCode uint32 `json:"code,omitempty"`
	Hash    string `json:"hash,omitempty"` // may be empty if it could not even deserialize our tx
	Message string `json:"message,omitempty"`
}

func (be BroadcastError) Error() string {
	return fmt.Sprintf("broadcast error: code = %d, hash = %s, msg = %s", be.ErrCode, be.Hash, be.Message)
}
