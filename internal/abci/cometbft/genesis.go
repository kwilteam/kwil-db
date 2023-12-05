package cometbft

import (
	"path/filepath"
)

// NOTE: we will soon be passing the genesis doc in memory rather than file.

// CometBFT file and folder names. These will be under the chain root directory.
// e.g. With "abci" a the chain root directory set in cometbft's config,
// this give the paths "abci/config/genesis.json" and "abci/data".
const (
	DataDir          = "data"
	GenesisJSONName  = "genesis.json"
	AddrBookFileName = "addrbook.json"
)

func AddrBookPath(chainRootDir string) string {
	return filepath.Join(chainRootDir, AddrBookFileName)
}
