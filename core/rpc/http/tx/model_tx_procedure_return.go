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
	IsTable bool `json:"is_table,omitempty"`
	Fields []TxTypedVariable `json:"fields,omitempty"`
}
