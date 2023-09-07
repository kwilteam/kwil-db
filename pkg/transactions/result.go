package transactions

// TransactionResult is the result of a transaction execution on chain
type TransactionResult struct {
	Code      uint32   `json:"code,omitempty"`
	Log       string   `json:"log,omitempty"`
	GasUsed   int64    `json:"gas_used,omitempty"`
	GasWanted int64    `json:"gas_wanted,omitempty"`
	Data      []byte   `json:"data,omitempty"`
	Events    [][]byte `json:"events,omitempty"`
}