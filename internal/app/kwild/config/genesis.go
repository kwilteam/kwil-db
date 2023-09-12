package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kwilteam/kwil-db/pkg/abci/cometbft"

	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
)

const (
	ABCIPubKeyTypeEd25519 = "ed25519"
	chainIDPrefix         = "kwil-chain-"
)

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
	ConsensusParams *ConsensusParams   `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator `json:"validators,omitempty"`
}

type GenesisValidator struct {
	Address string `json:"address"`
	PubKey  []byte `json:"pub_key"`
	Power   int64  `json:"power"`
	Name    string `json:"name"`
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

	// Number of blocks after which the validators join request expires if not approved
	JoinExpiry int64 `json:"join_expiry"`
}

type VersionParams struct {
	App uint64 `json:"app"`
}

type GenesisParams struct {
	JoinExpiry      int64
	WithoutGasCosts bool
	WithoutNonces   bool
	ChainIDPrefix   string
}

func DefaultGenesisParams() *GenesisParams {
	return &GenesisParams{
		JoinExpiry:      86400,
		WithoutGasCosts: true,
		WithoutNonces:   false,
		ChainIDPrefix:   "kwil-chain-",
	}
}

// Default ConsensusParams
func defaultConsensusParams() *ConsensusParams {
	return &ConsensusParams{
		Block: BlockParams{
			MaxBytes: 22020096, // 21MB
			MaxGas:   -1,
		},

		Evidence: EvidenceParams{
			MaxAgeNumBlocks: 100000,         // 27.8 hrs at 1block/s
			MaxAgeDuration:  48 * time.Hour, // 2 days
			MaxBytes:        1048576,        // 1 MB
		},
		Version: VersionParams{
			App: 0,
		},

		Validator: ValidatorParams{
			PubKeyTypes: []string{ABCIPubKeyTypeEd25519},
			JoinExpiry:  86400, // approx 1 day considering block rate of 1 block/s
		},

		WithoutNonces:   false,
		WithoutGasCosts: true,
	}
}

// Generate a genesis file with default configuration
func GenerateGenesisConfig(pkey []cmtEd.PrivKey, genParams *GenesisParams) *GenesisConfig {
	genVals := make([]GenesisValidator, len(pkey))
	for idx, key := range pkey {
		pub := key.PubKey()
		addr := pub.Address()
		val := GenesisValidator{
			Address: hex.EncodeToString(addr),
			PubKey:  pub.Bytes(),
			Power:   1,
			Name:    fmt.Sprint("validator-", idx),
		}
		genVals[idx] = val
	}

	genConf := GenesisConfig{
		ChainID:         cometbft.GenerateChainID(genParams.ChainIDPrefix),
		GenesisTime:     genesisTime(),
		ConsensusParams: defaultConsensusParams(),
		Validators:      genVals,
	}

	genConf.ConsensusParams.Validator.JoinExpiry = genParams.JoinExpiry
	genConf.ConsensusParams.WithoutGasCosts = genParams.WithoutGasCosts
	genConf.ConsensusParams.WithoutNonces = genParams.WithoutNonces

	return &genConf
}

// Write a genesis file to disk
func (genConf *GenesisConfig) SaveAs(file string) error {
	genDocBytes, err := json.MarshalIndent(genConf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, genDocBytes, 0644)
}

// Load a genesis file from disk and parse it into a GenesisConfig
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

func generatePrivateKeyFile(keyPath string) (cmtEd.PrivKey, error) {
	privKey := cmtEd.GenPrivKey()
	keyHex := hex.EncodeToString(privKey[:])
	return privKey, os.WriteFile(keyPath, []byte(keyHex), 0600)
}

func readPrivateKeyFile(keyPath string) (cmtEd.PrivKey, error) {
	privKeyHexB, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}
	privKeyHex := string(bytes.TrimSpace(privKeyHexB))
	privB, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("error decoding private key: %v", err)
	}
	return cmtEd.PrivKey(privB), nil
}

func readOrCreatePrivateKeyFile(keyPath string, autogen bool) (privKey cmtEd.PrivKey, newKey bool, err error) {
	if fileExists(keyPath) {
		privKey, err = readPrivateKeyFile(keyPath)
		return
	}
	if !autogen {
		err = fmt.Errorf("private key not found")
		return
	}

	privKey, err = generatePrivateKeyFile(keyPath)
	newKey = true
	return
}

// loadGenesisAndPrivateKey generates private key and genesis file if not exist
//
//   - If genesis file exists but not private key file, it will generate private
//     key and start the node as a non-validator.
//   - Otherwise, the genesis file is generated based on the private key and
//     starts the node as a validator.
func LoadGenesisAndPrivateKey(autoGen bool, privKeyPath, chainRootDir string, genParams *GenesisParams) (privKey cmtEd.PrivKey, genesisCfg *GenesisConfig, newKey, newGenesis bool, err error) {
	// Get private key:
	//  - if private key file exists, load it.
	//  - else if in autogen mode, generate private key and write to file.
	//  - else fail

	privKey, newKey, err = readOrCreatePrivateKeyFile(privKeyPath, autoGen)
	if err != nil {
		return
	}

	abciCfgDir := filepath.Join(chainRootDir, cometbft.ConfigDir)
	genFile := filepath.Join(abciCfgDir, cometbft.GenesisJSONName) // i.e. <root>/abci/config/genesis.json
	if fileExists(genFile) {
		fmt.Printf("Found genesis file %v\n", genFile)
		genesisCfg, err = LoadGenesisConfig(genFile)
		if err != nil {
			err = fmt.Errorf("error loading genesis file %s: %v", genFile, err)
			return
		}
	} else {
		if !autoGen {
			err = fmt.Errorf("genesis file not found: %s", genFile)
			return
		}

		if err = os.MkdirAll(abciCfgDir, 0755); err != nil {
			err = fmt.Errorf("error creating abci config dir %s: %v", abciCfgDir, err)
			return
		}

		genesisCfg = GenerateGenesisConfig([]cmtEd.PrivKey{privKey}, genParams)
		err = genesisCfg.SaveAs(genFile)
		if err != nil {
			err = fmt.Errorf("unable to write genesis file %s: %v", genFile, err)
			return
		}
		newGenesis = true
	}

	return
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

func GenerateChainID(prefix string) string {
	return prefix + cmtrand.Str(6)
}
