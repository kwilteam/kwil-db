// Package types contains the type used by the chain stat RPC client and servers.
package types

import "github.com/kwilteam/kwil-db/core/types"

type ChainTx struct {
	Hash     types.Hash      `json:"hash"`
	Height   int64           `json:"height"`
	Index    uint32          `json:"index"`
	Tx       []byte          `json:"tx"`
	TxResult *types.TxResult `json:"tx_result"`
}
