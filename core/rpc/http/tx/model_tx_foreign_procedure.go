/*
 * kwil/tx/v1/account.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type TxForeignProcedure struct {
	Name string `json:"name,omitempty"`
	Parameters []TxDataType `json:"parameters,omitempty"`
	ReturnTypes *TxProcedureReturn `json:"return_types,omitempty"`
}
