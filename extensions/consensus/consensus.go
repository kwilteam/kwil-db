// Package consensus is used to apply customized rules for activation of
// hardforks defined in the genesis file.
//
// The goal of this as an extension is to minimize the need to modify internal
// kwild code to define coordinated changes to certain common or likely
// consensus rules.
//
// main capabilities:
//   - new transaction payload types, or payload versions
//   - modifications (add/remove/update) to the registered authenticators
//   - modifications to registered resolutions (governance-initiated transactions)
//   - configurable / pluggable serialization scheme
//   - consensus engine parameter updates, such as block size or other limits
//   - one time actions at activation height (see ethereum's TheDAO state changes),
//     like transferring or minting tokens, decreed validator set updates, special
//     dataset modifications - resolution-like actions that must be in tandem with
//     code and app logic change.
//
// The genesis files forks section may reference canonical forks defined in
// kwild code with named helpers like IsMyLogicV2(height uint64) switching
// between logic at any location. That requires kwild code updates; this
// consensus extension package aims to allow definition of custom hard fork
// aliases, that may modify the logic in certain well-defined ways. Arbitrary
// changes to kwild consensus logic may require changes to kwild internals (as
// is the standard approach in most blockchains), or adding new capabilities to
// this consensus extensions package and how kwild uses it to facilitate new
// types of live changes.
package consensus

import (
	"context"
	"fmt"
	"slices"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	authExt "github.com/kwilteam/kwil-db/extensions/auth"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
)

// Need fork order defined for unambiguous application of many at same height,
// such as is common with genesis on a new network.

// Hardforks contains all known hardfork definitions.  Include both
// canonical/kwild and extensions? This is just a list. When kwild loads
// genesis.json, it locates the named hardforks in this registry.
var Hardforks = map[string]*Hardfork{}

// RegisterHardfork registers a hardfork definition by the name that should be
// expected in the genesis file. The payload type is registered in the
// transaction package, but the route is enabled in the tx app at activation.
func RegisterHardfork(hf *Hardfork) {
	if _, have := Hardforks[hf.Name]; have {
		panic(fmt.Sprintf("already have hardfork %q", hf.Name))
	}
	Hardforks[hf.Name] = hf

	for _, newPayload := range hf.TxPayloads {
		transactions.RegisterPayload(newPayload.Type)
		// NOTE: newPayload.Route stays in the []Hardfork.TxPayloads until
		// activation, when TxApp/ABCI enable the tx route.
	}
}

// Hardfork specifies the fundamental changes affected by a named hardfork. If a
// field is nil or the zero value for the type, that particular change is not
// part of the hardfork's definition. When the hardfork is registered, only the
// defined changes are applied in the relevant parts of kwild.
//
// An instance of this type is designed initially for the simplest coordination,
// height, using the common/chain package to specify activation height. However,
// more generally, the Hardfork struct exposes a set of well-defined types of
// changes that may be implemented via the extension system *without modifying
// kwild code*. Such a change may be part of a resolution approved with the
// voting system, in which case it would be applied when threshold is reached
// rather than when an activation height is reached.
type Hardfork struct {
	// Name is the hard fork's code name.
	Name string
	// NOTE: Activation height is specified by genesis.json (or other dynamic
	// methods like signaling / voting). This struct defines the changes.

	// Optional consensus logic overrides below.  All of these will require
	// dynamic kwil-db support, such as registering new encoders and transaction
	// payloads, submitting consensus parameter updates to the consensus engine,
	// applying state adjustments at activation height in ABCI / TxApp, etc.

	// CATEGORY 1: registered functionality changes: payloads, authenticators,
	// resolutions, and data serialization schema (codecs).

	// TxPayloads specifies new transaction payload to recognize at activation.
	// To modify an existing payload in a backward incompatible way, instead
	// create a new version of the payload such as PayloadTypeExecuteActionV2.
	// Any such changes would also be accompanied by tooling updates.
	// TODO: consider payload removal/invalidation and replace.
	TxPayloads []Payload // Type() and tx app Route() implementation

	// AuthUpdates are updates (add/remove/change) to known Authenticators for
	// signature verification.
	AuthUpdates []*AuthMod

	// ResolutionUpdates are coordinated changes to the resolutions extension.
	// They may be added, modified (redefined), or removed at activation. Any
	// have the potential to break consensus, and should be done with a Hardfork.
	ResolutionUpdates []*ResolutionMod

	// Encoders are new encoder types to register *at activation height* for
	// core/types/serialize.Encode/Decode e.g. Borsch instead of RLP.
	// EncodingTypeCustom offset should be used as the first possible external
	// codec's type to leave space for more kwild canonical codecs in the
	// future. Choose an encoding type that does not collide with other codecs.
	// Any such changes would also be accompanied by tooling updates.
	Encoders []*serialize.Codec

	// CATEGORY 2: One time updates.

	// ParamsUpdates are updates to the consensus engine parameters that should
	// go into affect at after the activation height (returned to the consensus
	// engine when finalizing the block at this height). For example, a block
	// size change.
	ParamsUpdates *ParamUpdates

	// StateMod is triggered at activation. It can do anything, one time. For
	// instance, arbitrary change to application state via the Engine or more
	// directly to the DB may be made at the end of block at the activation
	// height. This is to be called inside the outer transaction of activation
	// block, so changes to state are captured in the normal apphash diff. This
	// is a reasonable capability for a hardfork to make state changes outside
	// of transaction execution, but most such changes can probably be achieved
	// through the resolution system and voting. Doing it in a hardfork would be
	// needed if there are other changes (either baked in to kwild or in the
	// above fields) that should be done in concert with the fork.
	StateMod func(context.Context, *common.App) error
}

// ParamUpdates is much like common/chain.BaseConsensusParams, but uses
// pointer fields since updates are typically sparse.
type ParamUpdates struct {
	Block     *chain.BlockParams     `json:"block,omitempty"`
	Evidence  *chain.EvidenceParams  `json:"evidence,omitempty"`
	Version   *chain.VersionParams   `json:"version,omitempty"`
	Validator *chain.ValidatorParams `json:"validator,omitempty"`
	Votes     *chain.VoteParams      `json:"votes,omitempty"`
	ABCI      *chain.ABCIParams      `json:"abci,omitempty"`
}

// ResolutionMod defines a modification to a consensus resolution used by the
// oracle system. A modification may be adding a new resolution, or modifying or
// removing an existing resolution.
type ResolutionMod struct {
	Name      string
	Operation resolutions.ModOperation
	Config    *resolutions.ResolutionConfig
}

// AuthMod defines a modification to an authenticator used to verify signatures.
// A modification may be adding a new authenticator, or modifying or removing an
// existing authenticator.
type AuthMod struct {
	Name      string
	Operation authExt.ModOperation
	Authn     auth.Authenticator
}

// MergeConsensusUpdates combines the specified parameter updates. Both inputs
// represent *updates* rather than the current set of active parameters, and any
// fields of each input may be nil. Each value is only updated if it is not the
// zero value. Any exceptions should be noted.
func MergeConsensusUpdates(params, update *ParamUpdates) {
	if update == nil {
		return
	}

	if update.Block != nil {
		if update.Block.MaxBytes > 0 { //allow setting just MaxGas
			if params.Block == nil {
				params.Block = new(chain.BlockParams)
			}
			params.Block.MaxBytes = update.Block.MaxBytes
			params.Block.AbciBlockSizeHandling = update.Block.AbciBlockSizeHandling // must specify
		}
		if update.Block.MaxGas != 0 {
			if params.Block == nil {
				params.Block = new(chain.BlockParams)
			}
			params.Block.MaxGas = update.Block.MaxGas
		}
	}
	if update.Evidence != nil {
		if update.Evidence.MaxAgeNumBlocks > 0 {
			if params.Evidence == nil {
				params.Evidence = new(chain.EvidenceParams)
			}
			params.Evidence.MaxAgeNumBlocks = update.Evidence.MaxAgeNumBlocks
		}
		if update.Evidence.MaxAgeDuration > 0 {
			if params.Evidence == nil {
				params.Evidence = new(chain.EvidenceParams)
			}
			params.Evidence.MaxAgeDuration = update.Evidence.MaxAgeDuration
		}
		if update.Evidence.MaxBytes > 0 {
			if params.Evidence == nil {
				params.Evidence = new(chain.EvidenceParams)
			}
			params.Evidence.MaxBytes = update.Evidence.MaxBytes
		}
	}
	if update.Validator != nil {
		if len(update.Validator.PubKeyTypes) > 0 {
			if params.Validator == nil {
				params.Validator = new(chain.ValidatorParams)
			}
			params.Validator.PubKeyTypes = slices.Clone(update.Validator.PubKeyTypes)
			params.Validator.JoinExpiry = update.Validator.JoinExpiry
		}
	}
	if update.Votes != nil {
		if params.Votes == nil {
			params.Votes = new(chain.VoteParams)
		}
		params.Votes.VoteExpiry = update.Votes.VoteExpiry
	}
	if update.Version != nil {
		if update.Version.App > 0 {
			if params.Version == nil {
				params.Version = new(chain.VersionParams)
			}
			params.Version.App = update.Version.App
		}
	}
	if update.ABCI != nil {
		if params.ABCI == nil {
			params.ABCI = new(chain.ABCIParams)
		}
		// NOTE: this will allow changing it from non-zero to zero. However,
		// vote extensions may not be disabled once enabled, so this should only
		// be done to "cancel" a planned upgrade that would enable them at a future height.
		params.ABCI.VoteExtensionsEnableHeight = update.ABCI.VoteExtensionsEnableHeight
	}
}
