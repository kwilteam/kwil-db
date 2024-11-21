package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

const (
	nodeDirPerm = 0755
)

// loadGenesisAndPrivateKey generates private key and genesis file if not exist
//
//   - If genesis file exists but not private key file, it will generate private
//     key and start the node as a non-validator.
//   - Otherwise, the genesis file is generated based on the private key and
//     starts the node as a validator.
func loadGenesisAndPrivateKey(autoGen bool, privKeyPath, rootDir string) (privKey *crypto.Ed25519PrivateKey, genesisCfg *chain.GenesisConfig, err error) {
	// Get private key:
	//  - if private key file exists, load it.
	//  - else if in autogen mode, generate private key and write to file.
	//  - else fail

	if err = os.MkdirAll(rootDir, nodeDirPerm); err != nil {
		return nil, nil, fmt.Errorf("failed to make root directory: %w", err)
	}

	chainRootDir := ABCIDir(rootDir)
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
		genesisCfg, err = chain.LoadGenesisConfig(genFile)
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

	genesisCfg = chain.NewGenesisWithValidator(pub)
	err = genesisCfg.SaveAs(genFile)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to write genesis file %s: %v", genFile, err)
	}
	fmt.Printf("Generated new genesis file: %v\n", genFile)
	return privKey, genesisCfg, nil
}

func InitPrivateKeyAndGenesis(cfg *config.KwildConfig, autogen bool) (privateKey *crypto.Ed25519PrivateKey,
	genConfig *chain.GenesisConfig, err error) {
	return loadGenesisAndPrivateKey(autogen, cfg.AppConfig.PrivateKeyPath, cfg.RootDir)
}
