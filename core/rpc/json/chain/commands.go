package chain

import (
	"github.com/kwilteam/kwil-db/core/types"
)

type HealthRequest struct{}

type BlockRequest struct {
	Height int64 `json:"height"`
	// Hash is the block hash. If both Height and Hash are provided, hash will be used
	Hash types.Hash `json:"hash"`
	Raw  bool       `json:"raw"`
}

type BlockResultRequest struct {
	Height int64 `json:"height"`
	// Hash is the block hash. If both Height and Hash are provided, hash will be used
	Hash types.Hash `json:"hash"`
}

type TxRequest struct {
	Hash types.Hash `json:"hash"`
}

type GenesisRequest struct{}
type ConsensusParamsRequest struct{}
type ValidatorsRequest struct{}
type UnconfirmedTxsRequest struct {
	Limit int `json:"limit"`
}
