package cometbft

import (
	"path/filepath"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtTypes "github.com/cometbft/cometbft/types"

	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/extensions/consensus"
)

// CometBFT file and folder names. These will be under the chain root directory.
// e.g. With "abci" a the chain root directory set in cometbft's config,
// this give the paths "abci/config/genesis.json" and "abci/data".
const (
	DataDir          = "data"
	GenesisJSONName  = "genesis.json"
	ConfigTOMLName   = "config.toml"
	AddrBookFileName = "addrbook.json"
)

func AddrBookPath(chainRootDir string) string {
	return filepath.Join(chainRootDir, AddrBookFileName)
}

// ExtractConsensusParams creates cometbft's ConsensusParams from kwild's, which
// includes a subset of cometbft's and other parameters that pertain to the ABCI
// application (kwild) rather than the consensus engine (cometbft).
func ExtractConsensusParams(cp *chain.BaseConsensusParams) *cmtTypes.ConsensusParams {
	return &cmtTypes.ConsensusParams{
		Block: cmtTypes.BlockParams{
			MaxBytes: cp.Block.MaxBytes,
			MaxGas:   cp.Block.MaxGas,
		},
		Evidence: cmtTypes.EvidenceParams{
			MaxAgeNumBlocks: cp.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  cp.Evidence.MaxAgeDuration,
			MaxBytes:        cp.Evidence.MaxBytes,
		},
		Version: cmtTypes.VersionParams{
			App: cp.Version.App,
		},
		Validator: cmtTypes.ValidatorParams{
			PubKeyTypes: cp.Validator.PubKeyTypes,
		},
		ABCI: cmtTypes.ABCIParams{
			VoteExtensionsEnableHeight: cp.ABCI.VoteExtensionsEnableHeight,
		},
	}
}

// ParamUpdatesToComet converts the parameter updates to cometBFT's
func ParamUpdatesToComet(up *consensus.ParamUpdates) *cmtproto.ConsensusParams {
	var params cmtproto.ConsensusParams
	if up.Block != nil {
		params.Block = &cmtproto.BlockParams{
			MaxBytes: up.Block.MaxBytes,
			MaxGas:   up.Block.MaxGas,
		}
	}
	if up.Evidence != nil {
		params.Evidence = &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: up.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  up.Evidence.MaxAgeDuration,
			MaxBytes:        up.Evidence.MaxBytes,
		}
	}
	if up.Version != nil {
		params.Version = &cmtproto.VersionParams{
			App: up.Version.App,
		}
	}
	if up.Validator != nil {
		params.Validator = &cmtproto.ValidatorParams{
			PubKeyTypes: up.Validator.PubKeyTypes,
		}
	}
	// NOTE: comet doesn't have a Vote field, nor Validator.JoinExpiry, etc.
	// We handle those.
	if up.ABCI != nil {
		params.Abci = &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: up.ABCI.VoteExtensionsEnableHeight,
		}
	}
	return &params
}
