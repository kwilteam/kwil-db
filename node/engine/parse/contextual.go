package parse

import (
	"github.com/kwilteam/kwil-db/core/types"
)

// TODO: all of these should be moved to the engine.
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
	// block_timestamp is the unix timestamp of the block, set by the block proposer.
	BlockTimestamp = "block_timestamp"
	// authenticator provides information on the authenticator used to sign the transaction.
	Authenticator = "authenticator"
	// SessionVars are the session variables that are available in the engine.
	// It maps the variable name to its type.
	SessionVars = map[string]*types.DataType{
		CallerVar:      types.TextType,
		TxidVar:        types.TextType,
		SignerVar:      types.BlobType,
		HeightVar:      types.IntType,
		ForeignCaller:  types.TextType,
		BlockTimestamp: types.IntType,
		Authenticator:  types.TextType,
	}
)
