package config

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	types "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/random"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

const (
	abciPubKeyTypeEd25519 = "ed25519"
	chainIDPrefix         = "kwil-chain-"

	nodeDirPerm = 0755
)

type HexBytes = types.HexBytes

type GenesisConfig struct {
	GenesisTime   time.Time `json:"genesis_time"`
	ChainID       string    `json:"chain_id"`
	InitialHeight int64     `json:"initial_height"`
	DataAppHash   []byte    `json:"app_hash"`

	/*
	 TODO: Can introduce app state later if needed. Used to specify raw initial state such as tokens etc,
	 abci init will generate a new app hash after applying this state
	*/
	// AppState json.RawMessage `json:"app_state"`
	ConsensusParams *ConsensusParams    `json:"consensus_params,omitempty"`
	Validators      []*GenesisValidator `json:"validators,omitempty"`
}

type GenesisValidator struct {
	PubKey HexBytes `json:"pub_key"`
	Power  int64    `json:"power"`
	Name   string   `json:"name"`
}

type ConsensusParams struct {
	Block     BlockParams     `json:"block"`
	Evidence  EvidenceParams  `json:"evidence"`
	Version   VersionParams   `json:"version"`
	Validator ValidatorParams `json:"validator"`

	WithoutNonces   bool `json:"without_nonces"`
	WithoutGasCosts bool `json:"without_gas_costs"`
}

type BlockParams struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
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

type VersionParams struct {
	App uint64 `json:"app"`
}

func generateChainID(prefix string) string {
	return prefix + random.String(8)
}

// DefaultGenesisConfig returns a new instance of a GenesisConfig with the
// default values set, which in particular includes no validators and a nil
// appHash. The chain ID will semi-random, with the prefix "kwil-chain-"
// followed by random alphanumeric characters.
func DefaultGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		GenesisTime:     genesisTime(),
		ChainID:         generateChainID(chainIDPrefix),
		InitialHeight:   1,
		DataAppHash:     nil,
		Validators:      nil,
		ConsensusParams: defaultConsensusParams(),
	}
}

func defaultConsensusParams() *ConsensusParams {
	return &ConsensusParams{
		Block: BlockParams{
			MaxBytes: 6 * 1024 * 1024, // 21 MiB
			MaxGas:   -1,
		},
		Evidence: EvidenceParams{
			MaxAgeNumBlocks: 100_000,        // 27.8 hrs at 1block/s
			MaxAgeDuration:  48 * time.Hour, // 2 days
			MaxBytes:        1024 * 1024,    // 1 MiB
		},
		Version: VersionParams{
			App: 0,
		},
		Validator: ValidatorParams{
			PubKeyTypes: []string{abciPubKeyTypeEd25519},
			JoinExpiry:  14400, // approx 1 day considering block rate of 6 sec/s
		},
		WithoutNonces:   false,
		WithoutGasCosts: true,
	}
}

// SaveAs writes the genesis config to a file.
func (genConfig *GenesisConfig) SaveAs(file string) error {
	genDocBytes, err := json.MarshalIndent(genConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, genDocBytes, 0644)
}

// LoadGenesisConfig loads a genesis file from disk and parse it into a
// GenesisConfig.
func LoadGenesisConfig(file string) (*GenesisConfig, error) {
	genConfig := &GenesisConfig{}
	genDocBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(genDocBytes, genConfig)
	if err != nil {
		return nil, err
	}
	return genConfig, nil
}

// loadGenesisAndPrivateKey generates private key and genesis file if not exist
//
//   - If genesis file exists but not private key file, it will generate private
//     key and start the node as a non-validator.
//   - Otherwise, the genesis file is generated based on the private key and
//     starts the node as a validator.
func loadGenesisAndPrivateKey(autoGen bool, privKeyPath, rootDir string) (privKey *crypto.Ed25519PrivateKey, genesisCfg *GenesisConfig, err error) {
	// Get private key:
	//  - if private key file exists, load it.
	//  - else if in autogen mode, generate private key and write to file.
	//  - else fail

	if err = os.MkdirAll(rootDir, nodeDirPerm); err != nil {
		return nil, nil, fmt.Errorf("failed to make root directory: %w", err)
	}

	chainRootDir := filepath.Join(rootDir, ABCIDirName)
	priv, pub, newKey, err := ReadOrCreatePrivateKeyFile(privKeyPath, autoGen)
	if err != nil {
		return nil, nil, err
	}
	if newKey {
		fmt.Printf("Generated new private key, path: %v\n", privKeyPath)
	}
	privKey, err = crypto.Ed25519PrivateKeyFromBytes(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private key: %v", err)
	}

	genFile := filepath.Join(rootDir, cometbft.GenesisJSONName)

	if fileExists(genFile) {
		genesisCfg, err = LoadGenesisConfig(genFile)
		if err != nil {
			return nil, nil, fmt.Errorf("error loading genesis file %s: %v", genFile, err)
		}
		return privKey, genesisCfg, nil
	}

	if !autoGen {
		return nil, nil, fmt.Errorf("genesis file not found: %s", genFile)
	}

	if err = os.MkdirAll(chainRootDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("error creating abci config dir %s: %v", chainRootDir, err)
	}

	genesisCfg = NewGenesisWithValidator(pub)
	err = genesisCfg.SaveAs(genFile)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to write genesis file %s: %v", genFile, err)
	}
	fmt.Printf("Generated new genesis file: %v\n", genFile)
	return privKey, genesisCfg, nil
}

func genesisTime() time.Time {
	return time.Now().Round(0).UTC()
}

/*
AppHash: App hash in the genesis file corresponds to the initial database state.

CometBFT internally hashes specific fields from the ConsensusParams from the Genesis,
but doesn't automatically validates the rest of the fields.

computeGenesisHash constructs app hash based on the fields introduced by the application
in the genesis file which aren't monitored by cometBFT for consensus purposes.

This app hash is used by the ABCI application to initialize the blockchain.

Currently includes:
  - AppHash (Database state)
  - Join Expiry
  - Without Gas Costs
  - Without Nonces
*/
func (genConf *GenesisConfig) ComputeGenesisHash() []byte {
	hasher := sha256.New()
	hasher.Write(genConf.DataAppHash)
	binary.Write(hasher, binary.LittleEndian, genConf.ConsensusParams.Validator.JoinExpiry)

	if genConf.ConsensusParams.WithoutGasCosts {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}

	if genConf.ConsensusParams.WithoutNonces {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}

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

func listFilesAlphabetically(filePath string) ([]string, error) {
	files, err := filepath.Glob(filePath)
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	return files, nil
}

func hashFile(filePath string, hasher hash.Hash) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(hasher, file)
	return err
}

func setGenesisAppHash(appHash []byte, genesisFile string) error {
	genesisConf, err := LoadGenesisConfig(genesisFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis file: %w", err)
	}

	genesisConf.DataAppHash = appHash
	if err := genesisConf.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("failed to save genesis file: %w", err)
	}
	return nil
}

// PatchGenesisAppHash computes the apphash from a full contents of all sqlite
// files in the provided folder, and if genesis file is provided, updates the
// app_hash in the file.
func PatchGenesisAppHash(sqliteDbDir, genesisFile string) ([]byte, error) {
	di, err := os.Stat(sqliteDbDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sqlite path: %v", err)
	}
	if !di.IsDir() {
		return nil, fmt.Errorf("sqlite path is not a directory: %v", sqliteDbDir)
	}
	// List all sqlite files in the given dir in lexicographical order
	files, err := listFilesAlphabetically(filepath.Join(sqliteDbDir, "*.sqlite"))
	if err != nil {
		return nil, err
	}
	// Allow len(files) == 0 ?

	// Generate DB Hash
	hasher := sha256.New()
	for _, file := range files {
		if err = hashFile(file, hasher); err != nil {
			return nil, err
		}
	}
	genesisHash := hasher.Sum(nil)

	// Optionally update the app_hash in the genesis file.
	if genesisFile != "" {
		err = setGenesisAppHash(genesisHash, genesisFile)
		if err != nil {
			return nil, err
		}
	}

	return genesisHash, nil
}
