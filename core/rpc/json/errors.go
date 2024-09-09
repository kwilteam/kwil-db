package jsonrpc

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
	ErrorTimeout        ErrorCode = -32001

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

	ErrorCallChallengeNotFound ErrorCode = -1001
	ErrorInvalidCallChallenge  ErrorCode = -1002
	ErrorCallChallengeExpired  ErrorCode = -1003
	ErrorInvalidCallSignature  ErrorCode = -1004
	ErrorMismatchCallAuthType  ErrorCode = -1005
	ErrorTooFastChallengeReqs  ErrorCode = -1006
	ErrorNoQueryWithPrivateRPC ErrorCode = -1007
)

// More detailed errors use a structured error type in the "data" field of the
// responses "error" object. These may include a code field for domain-specific
// codes.
