/*
 * kwil/tx/v1/account.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type TxProcedureReturn struct {
	IsTable bool `json:"isTable,omitempty"`
	Columns []TxTypedVariable `json:"columns,omitempty"`
}
