package chain

import (
	"github.com/kwilteam/kwil-db/core/types"
	chaintypes "github.com/kwilteam/kwil-db/core/types/chain"
)

// HealthResponse is the health check response.
type HealthResponse struct {
	ChainID string `json:"chain_id"`
	Height  int64  `json:"height"`
	Healthy bool   `json:"healthy"`
}

// BlockResponse is the block information
type BlockResponse chaintypes.Block

type BlockResultResponse chaintypes.BlockResult

type TxResponse chaintypes.Tx

// GenesisResponse is the same as kwil-db/config.GenesisConfig, with JSON tags.
type GenesisResponse chaintypes.Genesis

type ConsensusParamsResponse = types.NetworkParameters

/*func (r ConsensusParamsResponse) MarshalJSON() ([]byte, error) {
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
}*/

type ValidatorsResponse struct {
	Height     int64              `json:"height"`
	Validators []*types.Validator `json:"validators"`
}

type UnconfirmedTxsResponse struct {
	Total int                  `json:"total"`
	Txs   []chaintypes.NamedTx `json:"txs"`
}
