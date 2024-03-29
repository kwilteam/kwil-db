/*
 * kwil/tx/v1/account.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type TxForeignKey struct {
	ChildKeys []string `json:"child_keys,omitempty"`
	ParentKeys []string `json:"parent_keys,omitempty"`
	ParentTable string `json:"parent_table,omitempty"`
	Actions []TxForeignKeyAction `json:"actions,omitempty"`
}
