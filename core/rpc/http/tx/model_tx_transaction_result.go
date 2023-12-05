/*
 * kwil/tx/v1/account.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type TxTransactionResult struct {
	Code int64 `json:"code,omitempty"`
	Log string `json:"log,omitempty"`
	GasUsed string `json:"gas_used,omitempty"`
	GasWanted string `json:"gas_wanted,omitempty"`
	// Data contains the output of the transaction.
	Data string `json:"data,omitempty"`
	Events []string `json:"events,omitempty"`
}
