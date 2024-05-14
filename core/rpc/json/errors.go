package jsonrpc

import "fmt"

type ErrorCode int32

const (
	// JSON-RPC 2.0 spec errors
	ErrorParse          ErrorCode = -32700 // Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.
	ErrorInternal       ErrorCode = -32603 // Internal JSON-RPC error.
	ErrorInvalidParams  ErrorCode = -32602 // Invalid method parameter(s).
	ErrorUnknownMethod  ErrorCode = -32601 // The method does not exist / is not available
	ErrorInvalidRequest ErrorCode = -32600 // The JSON sent is not a valid Request object

	// implementation specific server errors may be on -32000 to -32099.

	// ErrorResultEncoding is when the application handles the request without
	// error, but a result structure fails to encode to JSON.
	ErrorResultEncoding ErrorCode = -32000

	// Application errors get the rest of the code space.

	ErrorTxInternal       ErrorCode = -200 // any issue from txApp, cometbft, etc. in handling a tx
	ErrorTxExecFailure    ErrorCode = -201 // txCode != transactions.CodeOk
	ErrorTxNotFound       ErrorCode = -202 // abci.ErrTxNotFound
	ErrorTxPayloadInvalid ErrorCode = -203

	ErrorEngineInternal        ErrorCode = -300
	ErrorEngineDatasetNotFound ErrorCode = -301
	ErrorEngineDatasetExists   ErrorCode = -302
	ErrorEngineInvalidSchema   ErrorCode = -303

	ErrorDBInternal ErrorCode = -400

	ErrorAccountInternal ErrorCode = -500

	ErrorIdentInternal ErrorCode = -600
	ErrorIdentInvalid  ErrorCode = -601

	ErrorNodeInternal ErrorCode = -700

	ErrorValidatorsInternal ErrorCode = -800
	ErrorValidatorNotFound  ErrorCode = -801

	// reserve -900 to -999 for the KGW
	ErrorKGWInternal         ErrorCode = -900
	ErrorKGWNotAuthorized    ErrorCode = -901
	ErrorKGWInvalidPayload   ErrorCode = -902
	ErrorKGWNotAllowed       ErrorCode = -903
	ErrorKGWNotFound         ErrorCode = -904
	ErrorKGWTooManyRequests  ErrorCode = -905
	ErrorKGWMethodNotAllowed ErrorCode = -906
)

// More detailed errors use a structured error type in the "data" field of the
// responses "error" object. These may include a code field for domain-specific
// codes.

// BroadcastError is a structured error object used by MethodBroadcast when
// creating a Response with this in Error.Data. This error type would typically
// be in an response with the ErrorTxExecFailure RPC ErrorCode.
type BroadcastError struct {
	// TxCode corresponds to a transactions.TxCode, rather than an RPC error code.
	TxCode  uint32 `json:"tx_code,omitempty"`
	Hash    string `json:"hash,omitempty"` // may be empty if it could not even deserialize our tx
	Message string `json:"message,omitempty"`
}

func (be BroadcastError) Error() string {
	return fmt.Sprintf("broadcast error: code = %d, hash = %s, msg = %s", be.TxCode, be.Hash, be.Message)
}
