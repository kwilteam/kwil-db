// Package types contains the type used by the chain RPC client and server.
package types

import (
	"encoding/json"

	"github.com/kwilteam/kwil-db/core/types"
)

type Tx struct {
	Hash     types.Hash         `json:"hash"`
	Height   int64              `json:"height"`
	Index    uint32             `json:"index"`
	Tx       *types.Transaction `json:"tx"`
	TxResult *types.TxResult    `json:"tx_result"`
}

type BlockHeader types.BlockHeader

func (b BlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Version     uint16     `json:"version"`
		Height      int64      `json:"height"`
		NumTxns     uint32     `json:"num_txns"`
		PrevHash    types.Hash `json:"prev_hash"`
		PrevAppHash types.Hash `json:"prev_app_hash"`
		// Timestamp is the unix millisecond timestamp
		Timestamp        int64      `json:"timestamp"`
		MerkleRoot       types.Hash `json:"merkle_root"`
		ValidatorSetHash types.Hash `json:"validator_set_hash"`
	}{
		Version:          b.Version,
		Height:           b.Height,
		NumTxns:          b.NumTxns,
		PrevHash:         b.PrevHash,
		PrevAppHash:      b.PrevAppHash,
		Timestamp:        b.Timestamp.UnixMilli(),
		MerkleRoot:       b.MerkleRoot,
		ValidatorSetHash: b.ValidatorSetHash,
	})
}

type Block struct {
	Header    *BlockHeader         `json:"header"`
	Txns      []*types.Transaction `json:"txns"`
	Signature []byte               `json:"signature"`
	Hash      types.Hash           `json:"hash"`
	AppHash   types.Hash           `json:"app_hash"`
}

type BlockResult struct {
	Height    int64            `json:"height"`
	Hash      types.Hash       `json:"hash"`
	TxResults []types.TxResult `json:"tx_results"`
}

type Genesis struct {
	ChainID string `json:"chain_id"`
	// Leader is the leader's public key.
	Leader types.HexBytes `json:"leader"`
	// Validators is the list of genesis validators (including the leader).
	Validators []*types.Validator `json:"validators"`
	// MaxBlockSize is the maximum size of a block in bytes.
	MaxBlockSize int64 `json:"max_block_size"`
	// JoinExpiry is the number of blocks after which the validators
	// join request expires if not approved.
	JoinExpiry int64 `json:"join_expiry"`
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool `json:"disabled_gas_costs"`
	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64 `json:"max_votes_per_tx"`
}

type NamedTx struct {
	Hash types.Hash         `json:"hash"`
	Tx   *types.Transaction `json:"tx"`
}
