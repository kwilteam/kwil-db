package jsonrpc

// Method is a type used for all recognized JSON-RPC method names.
type Method string

// server will
//  - dispatch to a registered handler by method name
//  - unmarshal params into an instance of the type for method

const (
	MethodUserVersion Method = "user.version"
	MethodPing        Method = "user.ping"
	MethodChainInfo   Method = "user.chain_info"
	MethodAccount     Method = "user.account"
	MethodBroadcast   Method = "user.broadcast"
	MethodCall        Method = "user.call"
	MethodDatabases   Method = "user.databases"
	MethodPrice       Method = "user.estimate_price"
	MethodQuery       Method = "user.query"
	MethodTxQuery     Method = "user.tx_query"
	MethodSchema      Method = "user.schema"
)
