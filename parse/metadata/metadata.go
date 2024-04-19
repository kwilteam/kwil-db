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
	// PgSessionVars are the session variables that are available in the engine.
	// It maps the variable name to its type.
	PgSessionVars = map[string]*types.DataType{
		CallerVar: types.TextType,
		TxidVar:   types.TextType,
	}
)
