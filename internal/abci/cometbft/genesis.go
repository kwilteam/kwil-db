package cometbft

import (
	"path/filepath"
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
