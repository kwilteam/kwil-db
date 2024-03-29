package cometbft

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/internal/utils"
)

// ResetState removes address book files plus all blockchain databases.
// It always returns nil error, only printing errors. (?)
func ResetState(chainRootDir string) error {
	chainDBDir := filepath.Join(chainRootDir, DataDir)

	blockdb := filepath.Join(chainDBDir, "blockstore.db")
	state := filepath.Join(chainDBDir, "state.db")
	wal := filepath.Join(chainDBDir, "cs.wal")
	evidence := filepath.Join(chainDBDir, "evidence.db")
	txIndex := filepath.Join(chainDBDir, "tx_index.db")

	// Why don't we just delete chainDBDir like Reset() does?

	if utils.FileExists(blockdb) {
		if err := os.RemoveAll(blockdb); err == nil {
			fmt.Println("Removed all blockstore.db", "dir", blockdb)
		} else {
			fmt.Println("error removing all blockstore.db", "dir", blockdb, "err", err)
		}
	}

	if utils.FileExists(state) {
		if err := os.RemoveAll(state); err == nil {
			fmt.Println("Removed all state.db", "dir", state)
		} else {
			fmt.Println("error removing all state.db", "dir", state, "err", err)
		}
	}

	if utils.FileExists(wal) {
		if err := os.RemoveAll(wal); err == nil {
			fmt.Println("Removed all cs.wal", "dir", wal)
		} else {
			fmt.Println("error removing all cs.wal", "dir", wal, "err", err)
		}
	}

	if utils.FileExists(evidence) {
		if err := os.RemoveAll(evidence); err == nil {
			fmt.Println("Removed all evidence.db", "dir", evidence)
		} else {
			fmt.Println("error removing all evidence.db", "dir", evidence, "err", err)
		}
	}

	if utils.FileExists(txIndex) {
		if err := os.RemoveAll(txIndex); err == nil {
			fmt.Println("Removed all tx_index.db", "dir", txIndex)
		} else {
			fmt.Println("error removing all tx_index.db", "dir", txIndex, "err", err)
		}
	}

	return nil
}
