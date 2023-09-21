package cometbft

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtRand "github.com/cometbft/cometbft/libs/rand"
	cmtTypes "github.com/cometbft/cometbft/types"
	cmtTime "github.com/cometbft/cometbft/types/time"
)

// NOTE: we will soon be passing the genesis doc in memory rather than file.

// CometBFT file and folder names. These will be under the chain root directory.
// e.g. With "abci" a the chain root directory set in cometbft's config,
// this give the paths "abci/config/genesis.json" and "abci/data".
const (
	ConfigDir        = "config"
	DataDir          = "data"
	GenesisJSONName  = "genesis.json"
	AddrBookFileName = "addrbook.json"
)

// GenesisPath returns the path of the genesis file given a chain root
// directory. e.g. Given the path to the "abci" folder:
// <kwild_root/abci>/config/genesis.json
func GenesisPath(chainRootDir string) string {
	abciCfgDir := filepath.Join(chainRootDir, ConfigDir)
	return filepath.Join(abciCfgDir, GenesisJSONName)
}

func AddrBookPath(chainRootDir string) string {
	abciCfgDir := filepath.Join(chainRootDir, ConfigDir)
	return filepath.Join(abciCfgDir, AddrBookFileName)
}

func GeneratePrivateKeyFile(keyPath string) (cmtEd.PrivKey, error) {
	privKey := cmtEd.GenPrivKey()
	keyHex := hex.EncodeToString(privKey[:])
	return privKey, os.WriteFile(keyPath, []byte(keyHex), 0600)
}

func GenerateGenesisFile(path string, pkeys []cmtEd.PrivKey, chainIDPrefix string) error {
	doc := GenesisDoc(pkeys, chainIDPrefix)
	return doc.SaveAs(path)
}

func GenesisDocBytes(pkeys []cmtEd.PrivKey, chainIDPrefix string) ([]byte, error) {
	doc := GenesisDoc(pkeys, chainIDPrefix)
	return cmtjson.MarshalIndent(doc, "", "  ")
}

func GenesisDoc(pkeys []cmtEd.PrivKey, chainIDPrefix string) *cmtTypes.GenesisDoc {
	genVals := make([]cmtTypes.GenesisValidator, len(pkeys))
	for idx, key := range pkeys {
		pub := key.PubKey().(cmtEd.PubKey)
		addr := pub.Address()
		val := cmtTypes.GenesisValidator{
			Address: addr,
			PubKey:  pub,
			Power:   1,
			Name:    fmt.Sprint("validator-", idx),
		}
		genVals[idx] = val
	}

	genDoc := cmtTypes.GenesisDoc{
		ChainID:         chainIDPrefix + cmtRand.Str(6),
		GenesisTime:     cmtTime.Now(),
		ConsensusParams: cmtTypes.DefaultConsensusParams(), // includes VoteExtensionsEnableHeight: 0, (disabled)
		Validators:      genVals,
	}
	return &genDoc
}

func SetGenesisAppHash(appHash []byte, genesisFile string) error {
	genesisDoc, err := cmtTypes.GenesisDocFromFile(genesisFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis file: %w", err)
	}
	genesisDoc.AppHash = appHash
	if err := genesisDoc.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("failed to save genesis file: %w", err)
	}
	return nil
}
