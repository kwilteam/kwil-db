// package metadata includes important metadata for the database engine.
// This includes supported functions, contextual variables, and more.
package metadata

import "github.com/kwilteam/kwil-db/core/types"

var (
	// PgSessionPrefix is the prefix for all session variables.
	// It is used in combination with Postgre's current_setting function
	// to set contextual variables.
	PgSessionPrefix = "ctx"
	// caller is the session variable for the caller.
	CallerVar = "caller"
	// txid is the session variable for the transaction id.
	TxidVar = "txid"
	// signer is the session variable for the signer.
	SignerVar = "signer"
	// PgSessionVars are the session variables that are available in the engine.
	// It maps the variable name to its type.
	PgSessionVars = map[string]*types.DataType{
		CallerVar: types.TextType,
		TxidVar:   types.TextType,
		SignerVar: types.BlobType,
	}
)

// GetSessionVariable checks if a variable is a session variable.
// If the input has an @, it will be removed.
func GetSessionVariable(name string) (*types.DataType, bool) {
	if len(name) == 0 {
		return nil, false
	}
	if name[0] == '@' {
		name = name[1:]
	}
	t, ok := PgSessionVars[name]
	return t, ok
}
