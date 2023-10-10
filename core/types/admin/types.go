// Package types contains the type used by the administrative RPC client and
// servers. These types are shared like the protobuf types that these mimick,
// but they contain the json tags and marshalling behavior we want, plus they
// are not burdened by fields and methods from the google protobuf packages.
package types

import (
	"time"

	"github.com/kwilteam/kwil-db/core/types"
)

// TODO(@jchappelow): doc these types.

type NodeInfo struct {
	ChainID         string `json:"chain_id"`
	Name            string `json:"name"`
	NodeID          string `json:"node_id"`
	ProtocolVersion uint64 `json:"proto_ver"`
	AppVersion      uint64 `json:"app_ver"`
	BlockVersion    uint64 `json:"block_ver"`
	ListenAddr      string `json:"listen_addr"`
	RPCAddr         string `json:"rpc_addr"`
}

type SyncInfo struct {
	AppHash         string    `json:"app_hash"`
	BestBlockHash   string    `json:"best_block_hash"`
	BestBlockHeight int64     `json:"best_block_height"`
	BestBlockTime   time.Time `json:"best_block_time"`

	Syncing bool `json:"syncing"`
}

type ValidatorInfo struct {
	PubKey     types.HexBytes `json:"pubkey"`
	PubKeyType string         `json:"pubkey_type"`
	Power      int64          `json:"power"`
}

type Status struct {
	Node      *NodeInfo      `json:"node"`
	Sync      *SyncInfo      `json:"sync"`
	Validator *ValidatorInfo `json:"current_validator"`
}

type PeerInfo struct {
	NodeInfo   *NodeInfo `json:"node"`
	Inbound    bool      `json:"inbound"`
	RemoteAddr string    `json:"remote_addr"`
}
