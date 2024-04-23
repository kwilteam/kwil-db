package jsonrpc

// Method is a type used for all recognized JSON-RPC method names.
type Method string

// server will
//  - dispatch to a registered handler by method name
//  - unmarshal params into an instance of the type for method

const (
	MethodPing      Method = "ping"
	MethodChainInfo Method = "chain_info"
	MethodAccount   Method = "account"
	MethodBroadcast Method = "broadcast"
	MethodCall      Method = "call"
	MethodDatabases Method = "databases"
	MethodPrice     Method = "estimate_price"
	MethodQuery     Method = "query"
	MethodTxQuery   Method = "tx_query"
	MethodSchema    Method = "schema"
)
