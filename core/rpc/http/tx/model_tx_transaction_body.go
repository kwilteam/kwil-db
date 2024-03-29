/*
 * kwil/tx/v1/account.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type TxTransactionBody struct {
	Payload string `json:"payload,omitempty"`
	PayloadType string `json:"payload_type,omitempty"`
	Fee string `json:"fee,omitempty"`
	Nonce string `json:"nonce,omitempty"`
	ChainId string `json:"chain_id,omitempty"`
	Description string `json:"description,omitempty"`
}
