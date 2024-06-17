package parse

import (
	"github.com/kwilteam/kwil-db/core/types"
)

var (
	// caller is the session variable for the caller.
	CallerVar = "caller"
	// txid is the session variable for the transaction id.
	TxidVar = "txid"
	// signer is the session variable for the signer.
	SignerVar = "signer"
	// height is the session variable for the block height.
	HeightVar = "height"
	// foreign_caller is the dbid of the schema that made a foreign call.
	ForeignCaller = "foreign_caller"
	// SessionVars are the session variables that are available in the engine.
	// It maps the variable name to its type.
	SessionVars = map[string]*types.DataType{
		CallerVar:     types.TextType,
		TxidVar:       types.TextType,
		SignerVar:     types.BlobType,
		HeightVar:     types.IntType,
		ForeignCaller: types.TextType,
	}
)

// makeSessionVars creates a new map of session variables.
// It includes the @ symbol in the keys.
func makeSessionVars() map[string]*types.DataType {
	newMap := make(map[string]*types.DataType)
	for k, v := range SessionVars {
		newMap["@"+k] = v.Copy()
	}
	return newMap
}
