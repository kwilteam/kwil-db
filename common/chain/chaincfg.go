// Package chain defines kwild's chain configuration types that model the
// genesis.json document. This is distinct from application runtime
// configuration (see config.toml) that does not affect consensus.
package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"sort"
	"time"

	"github.com/kwilteam/kwil-db/common/chain/forks"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/random"
)

const (
	abciPubKeyTypeEd25519 = "ed25519"
	chainIDPrefix         = "kwil-chain-"
)

type HexBytes = types.HexBytes

// GenesisConfig is the genesis chain configuration. Use LoadGenesisConfig to
// load from a JSON file and populate the ForkHeights field.
type GenesisConfig struct {
	GenesisTime   time.Time    `json:"genesis_time"`
	ChainID       string       `json:"chain_id"`
	InitialHeight int64        `json:"initial_height"`
	DataAppHash   []byte       `json:"app_hash"`
	Alloc         GenesisAlloc `json:"alloc,omitempty"`

	// ForkHeights is a map of named forks to activation heights. This is a map
	// to support forks defined by extensions. Use the Forks method to get a
	// Forks structure containing named fields for the canonical forks.
	ForkHeights map[string]*uint64 `json:"activations"` // e.g. {"activations": {"longhorn": 12220000}}

	ConsensusParams *ConsensusParams    `json:"consensus_params,omitempty"`
	Validators      []*GenesisValidator `json:"validators,omitempty"`
}

// Forks creates a forks.Forks instance from the ForkHeights map field.
func (gc *GenesisConfig) Forks() *forks.Forks {
	return forks.NewForks(gc.ForkHeights)
}

// Forks is a singleton instance of the hardforks parsed from a GenesisConfig.
// This is a convenience to provide global access to loaded hardfork
// configuration to other packages used in an application. The application
// should set this following LoadGenesisConfig, and prior to starting operations
// that rely on the config.
var Forks *forks.Forks

// SetForks initializes the package-level Forks variable.
func SetForks(forkHeights map[string]*uint64) {
	Forks = forks.NewForks(forkHeights)
}

type GenesisAlloc map[string]*big.Int

type GenesisValidator struct {
	PubKey HexBytes `json:"pub_key"`
	Power  int64    `json:"power"`
	Name   string   `json:"name"`
}

type BaseConsensusParams struct {
	Block     BlockParams     `json:"block"`
	Evidence  EvidenceParams  `json:"evidence"`
	Version   VersionParams   `json:"version"`
	Validator ValidatorParams `json:"validator"`
	Votes     VoteParams      `json:"votes"`
	ABCI      ABCIParams      `json:"abci"`
}

// ConsensusParams combines BaseConsensusParams with WithoutGasCosts.
type ConsensusParams struct {
	BaseConsensusParams

	// This is unchangeable after genesis.
	WithoutGasCosts bool `json:"without_gas_costs"`
}

type ABCIParams struct {
	VoteExtensionsEnableHeight int64 `json:"vote_extensions_enable_height"`
}

type BlockParams struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
	// AbciBlockSizeHandling indicates to give cometbft MaxBytes=-1 so it is the
	// ABCI application's job to respect MaxBytes when preparing block
	// proposals. If false, cometbft will limit the number of transactions
	// offered to ABCI.
	AbciBlockSizeHandling bool `json:"abci_max_bytes"` // if true, give -1 to consensus engine, and abci validator enforces instead
}

type EvidenceParams struct {
	MaxAgeNumBlocks int64         `json:"max_age_num_blocks"`
	MaxAgeDuration  time.Duration `json:"max_age_duration"`
	MaxBytes        int64         `json:"max_bytes"`
}

type ValidatorParams struct {
	PubKeyTypes []string `json:"pub_key_types"`

	// JoinExpiry is the number of blocks after which the validators join
	// request expires if not approved.
	JoinExpiry int64 `json:"join_expiry"`
}

type VoteParams struct {
	// VoteExpiry is the number of blocks after which the resolution expires if
	// consensus is not reached.
	VoteExpiry int64 `json:"vote_expiry"`
}

type VersionParams struct {
	App uint64 `json:"app"`
}

func defaultConsensusParams() *ConsensusParams {
	return &ConsensusParams{
		BaseConsensusParams: BaseConsensusParams{
			Block: BlockParams{
				// TODO: in an upgrade, set MaxBytes to -1 so we can do the
				// truncation in PrepareProposal after our other processing.
				MaxBytes:              6 * 1024 * 1024, // 6 MiB
				MaxGas:                -1,
				AbciBlockSizeHandling: false, // false means cometbft will limit txns provided to PrepareProposal
			},
			Evidence: EvidenceParams{
				MaxAgeNumBlocks: 100_000,        // 27.8 hrs at 1 block/s
				MaxAgeDuration:  48 * time.Hour, // 2 days
				MaxBytes:        1024 * 1024,    // 1 MiB
			},
			Version: VersionParams{
				App: 0,
			},
			Validator: ValidatorParams{
				PubKeyTypes: []string{abciPubKeyTypeEd25519},
				JoinExpiry:  14400, // approx 1 day considering block rate of 6 sec/blk
			},
			Votes: VoteParams{
				VoteExpiry: 14400, // approx 1 day considering block rate of 6 sec/blk
			},
			ABCI: ABCIParams{
				VoteExtensionsEnableHeight: 0, // disabled, needs coordinated upgrade to enable
			},
		},
		WithoutGasCosts: true,
	}
}

// DefaultGenesisConfig returns a new instance of a GenesisConfig with the
// default values set, which in particular includes no validators and a nil
// appHash. The chain ID will semi-random, with the prefix "kwil-chain-"
// followed by random alphanumeric characters.
func DefaultGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		GenesisTime:     time.Now().Round(0).UTC(),
		ChainID:         chainIDPrefix + random.String(8),
		InitialHeight:   1,
		DataAppHash:     nil,
		Validators:      nil,
		ConsensusParams: defaultConsensusParams(),
	}
}

// SaveAs writes the genesis config to a file.
func (gc *GenesisConfig) SaveAs(file string) error {
	genDocBytes, err := json.MarshalIndent(gc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, genDocBytes, 0644)
}

// LoadGenesisConfig loads a genesis file from disk and parse it into a
// GenesisConfig.
func LoadGenesisConfig(file string) (*GenesisConfig, error) {
	genDocBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	gc := &GenesisConfig{}
	err = json.Unmarshal(genDocBytes, gc)
	if err != nil {
		return nil, err
	}

	// Set the Forks singleton for package level access
	// Forks = gc.Forks()

	return gc, json.Unmarshal(genDocBytes, gc)
}

// ComputeGenesisHash constructs the app hash based on the fields introduced by
// the application in the genesis file which aren't monitored by cometBFT for
// consensus purposes.
//
// This app hash is used by the ABCI application to initialize the blockchain.
// The app hash in the genesis file corresponds to the initial database state.
//
// CometBFT internally hashes specific fields from the ConsensusParams, but
// doesn't automatically validates the rest of the fields.
//
// Currently includes:
//   - AppHash (Datastores state)
//   - Join Expiry
//   - Without Gas Costs
//   - Without Nonces
//   - Allocs (account allocations, same format as ethereum genesis.json)
//   - Vote Expiry
func (gc *GenesisConfig) ComputeGenesisHash() []byte {
	hasher := sha256.New()
	hasher.Write(gc.DataAppHash)
	binary.Write(hasher, binary.LittleEndian, gc.ConsensusParams.Validator.JoinExpiry)

	if gc.ConsensusParams.WithoutGasCosts {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}

	type genesisAlloc struct {
		acct string
		bal  *big.Int
	}
	allocs := make([]genesisAlloc, 0, len(gc.Alloc))
	for acct, bal := range gc.Alloc {
		allocs = append(allocs, genesisAlloc{
			acct: acct,
			bal:  bal,
		})
	}
	sort.Slice(allocs, func(i, j int) bool {
		return allocs[i].acct < allocs[j].acct
	})
	for _, alloc := range allocs {
		hasher.Write([]byte(alloc.acct))
		hasher.Write(alloc.bal.Bytes())
	}

	binary.Write(hasher, binary.LittleEndian, gc.ConsensusParams.Votes.VoteExpiry)

	// Note: Do not consider gc.Forks(): There is an upgrade window, where
	// software and genesis.json file updates may be applied prior to a deadline
	// when the change is active. These are operator configurable changes to
	// rules that are realized *after* genesis. Even with ComputeGenesisHash
	// called only in InitChain (genesis), this would still cause app hash
	// divergence when synchronizing a new node.

	return hasher.Sum(nil)
}

func NewGenesisWithValidator(pubKey []byte) *GenesisConfig {
	genesisCfg := DefaultGenesisConfig()
	const power = 1
	genesisCfg.Validators = append(genesisCfg.Validators, &GenesisValidator{
		PubKey: pubKey,
		Power:  power,
		Name:   "validator-0",
	})
	return genesisCfg
}
