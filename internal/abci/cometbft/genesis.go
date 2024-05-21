package cometbft

import (
	"path/filepath"

	cmtAPITypes "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmtTypes "github.com/cometbft/cometbft/types"
	gogotypes "github.com/cosmos/gogoproto/types" // c'mon

	"github.com/kwilteam/kwil-db/common"
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
// application (kwild) rather than the consensus engine (cometbft). The
// appVersion indicates state machine logic and it not an application parameter.
func ExtractConsensusParams(cp *chain.BaseConsensusParams, appVersion uint64) *cmtTypes.ConsensusParams {
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
			App: appVersion,
		},
		Validator: cmtTypes.ValidatorParams{
			PubKeyTypes: cp.Validator.PubKeyTypes,
		},
		Synchrony: cmtTypes.DefaultSynchronyParams(),
		Feature: cmtTypes.FeatureParams{
			VoteExtensionsEnableHeight: 0, // disabled for now
			// PbtsEnableHeight: ,
		},
	}
}

// MergeConsensusParams merges cometbft's ConsensusParams with kwild's NetworkParameters
// to create a unified representation of the chain's consensus parameters.
func MergeConsensusParams(cometbftParams *cmtTypes.ConsensusParams, abciParams *common.NetworkParameters) *chain.ConsensusParams {
	veeHeight := cometbftParams.Feature.VoteExtensionsEnableHeight
	pbsteHeight := cometbftParams.Feature.PbtsEnableHeight
	return &chain.ConsensusParams{
		BaseConsensusParams: chain.BaseConsensusParams{
			Block: chain.BlockParams{
				MaxBytes: abciParams.MaxBlockSize,
				MaxGas:   cometbftParams.Block.MaxGas,
			},
			Evidence: chain.EvidenceParams{
				MaxAgeNumBlocks: cometbftParams.Evidence.MaxAgeNumBlocks,
				MaxAgeDuration:  cometbftParams.Evidence.MaxAgeDuration,
				MaxBytes:        cometbftParams.Evidence.MaxBytes,
			},
			Validator: chain.ValidatorParams{
				PubKeyTypes: cometbftParams.Validator.PubKeyTypes,
				JoinExpiry:  abciParams.JoinExpiry,
			},
			Votes: chain.VoteParams{
				VoteExpiry:    abciParams.VoteExpiry,
				MaxVotesPerTx: abciParams.MaxVotesPerTx,
			},
			Feature: chain.FeatureParams{
				VoteExtensionsEnableHeight: &veeHeight,
				PbtsEnableHeight:           &pbsteHeight,
			},
			Synchrony: chain.SynchronyParams{},
		},
		WithoutGasCosts: abciParams.DisabledGasCosts,
	}
}

// ParamUpdatesToComet converts the parameter updates to cometBFT's
func ParamUpdatesToComet(up *consensus.ParamUpdates) *cmtAPITypes.ConsensusParams {
	var params cmtAPITypes.ConsensusParams
	if up.Block != nil {
		params.Block = &cmtAPITypes.BlockParams{
			MaxBytes: up.Block.MaxBytes,
			MaxGas:   up.Block.MaxGas,
		}
	}
	if up.Evidence != nil {
		params.Evidence = &cmtAPITypes.EvidenceParams{
			MaxAgeNumBlocks: up.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  up.Evidence.MaxAgeDuration,
			MaxBytes:        up.Evidence.MaxBytes,
		}
	}
	if up.Version != nil {
		params.Version = &cmtAPITypes.VersionParams{
			App: up.Version.App,
		}
	}
	if up.Validator != nil {
		params.Validator = &cmtAPITypes.ValidatorParams{
			PubKeyTypes: up.Validator.PubKeyTypes,
		}
	}
	// NOTE: comet doesn't have a Vote field, nor Validator.JoinExpiry, etc.
	// We handle those.
	if up.Feature != nil {
		params.Feature = new(cmtAPITypes.FeatureParams)
		if veeh := up.Feature.VoteExtensionsEnableHeight; veeh != nil {
			params.Feature.VoteExtensionsEnableHeight = &gogotypes.Int64Value{
				Value: *veeh,
			}
		}
		if pbsteh := up.Feature.PbtsEnableHeight; pbsteh != nil {
			params.Feature.PbtsEnableHeight = &gogotypes.Int64Value{
				Value: *pbsteh,
			}
		}
	}
	// TODO: sychrony
	return &params
}
