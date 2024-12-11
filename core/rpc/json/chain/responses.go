package chain

import (
	"encoding/json"

	"github.com/kwilteam/kwil-db/core/types"
)

// HealthResponse is the health check response.
type HealthResponse struct {
	ChainID string `json:"chain_id"`
	Height  int64  `json:"height"`
	Healthy bool   `json:"healthy"`
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

// BlockResponse is the block information
type BlockResponse struct {
	Header    *BlockHeader `json:"header"`
	Txns      [][]byte     `json:"txns"`
	Signature []byte       `json:"signature"`
	Hash      types.Hash   `json:"hash"`
	AppHash   types.Hash   `json:"app_hash"`
}

type BlockResultResponse struct {
	Height    int64            `json:"height"`
	TxResults []types.TxResult `json:"tx_results"`
}

// GenesisResponse is the same as kwil-db/config.GenesisConfig, with JSON tags.
type GenesisResponse struct {
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
	// VoteExpiry is the default number of blocks after which the validators
	// vote expires if not approved.
	VoteExpiry int64 `json:"vote_expiry"`
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool `json:"disabled_gas_costs"`
	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64 `json:"max_votes_per_tx"`
}

type ConsensusParamsResponse types.ConsensusParams

func (r ConsensusParamsResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		// MaxBlockSize is the maximum size of a block in bytes.
		MaxBlockSize int64 `json:"max_block_size"`
		// JoinExpiry is the number of blocks after which the validators
		// join request expires if not approved.
		JoinExpiry int64 `json:"join_expiry"`
		// VoteExpiry is the default number of blocks after which the validators
		// vote expires if not approved.
		VoteExpiry int64 `json:"vote_expiry"`
		// DisabledGasCosts dictates whether gas costs are disabled.
		DisabledGasCosts bool `json:"disabled_gas_costs"`

		// MigrationStatus determines the status of the migration.
		MigrationStatus string `json:"migration_status"`

		// MaxVotesPerTx is the maximum number of votes that can be included in a
		// single transaction.
		MaxVotesPerTx int64 `json:"max_votes_per_tx"`
	}{
		MaxBlockSize:     r.MaxBlockSize,
		JoinExpiry:       r.JoinExpiry,
		VoteExpiry:       r.VoteExpiry,
		DisabledGasCosts: r.DisabledGasCosts,
		MigrationStatus:  string(r.MigrationStatus),
		MaxVotesPerTx:    r.MaxVotesPerTx,
	})
}

type ValidatorsResponse struct {
	Height     int64              `json:"height"`
	Validators []*types.Validator `json:"validators"`
}

type NamedTx struct {
	Hash types.Hash `json:"hash"`
	Tx   []byte     `json:"tx"`
}

type UnconfirmedTxsResponse struct {
	Total int       `json:"total"`
	Txs   []NamedTx `json:"txs"`
}
