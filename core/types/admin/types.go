// Package types contains the type used by the administrative RPC client and
// servers. These types are shared like the protobuf types that these mimick,
// but they contain the json tags and marshalling behavior we want, plus they
// are not burdened by fields and methods from the google protobuf packages.
package types

import (
	"time"

	"github.com/kwilteam/kwil-db/core/types"
)

// NodeInfo describes a peer node. This may be a peer or a node being
// administered.
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

// SyncInfo describes the sync state of a node.
type SyncInfo struct {
	AppHash         string    `json:"app_hash"`
	BestBlockHash   string    `json:"best_block_hash"`
	BestBlockHeight int64     `json:"best_block_height"`
	BestBlockTime   time.Time `json:"best_block_time"`

	Syncing bool `json:"syncing"`
}

// ValidatorInfo describes a validator node.
type ValidatorInfo struct {
	PubKey types.HexBytes `json:"pubkey"`
	Power  int64          `json:"power"`
}

// type ValidatorInfo = types.Validator

// Status includes a comprehensive summary of a nodes status, including if the
// service is running, its best block and if it is syncing, its identity on
// the network, and the node's validator identity if it is one. Note that our
// validator is part of the node rather than an external signer.
type Status struct {
	Node      *NodeInfo      `json:"node"`
	Sync      *SyncInfo      `json:"sync"`
	Validator *ValidatorInfo `json:"validator"`
	Migration MigrationInfo  `json:"migration"`
}

// PeerInfo describes a connected peer node.
type PeerInfo struct {
	NodeInfo   *NodeInfo `json:"node"`
	Inbound    bool      `json:"inbound"`
	RemoteAddr string    `json:"remote_addr"`
}

type MigrationInfo struct {
	Status      string `json:"status"`
	StartHeight int64  `json:"start_height"`
	EndHeight   int64  `json:"end_height"`
}
