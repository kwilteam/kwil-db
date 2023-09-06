package cometbft

import (
	"fmt"
	"os"
	"path/filepath"

	cmtos "github.com/cometbft/cometbft/libs/os"
)

// ResetState removes address book files plus all blockchain databases.
func ResetState(chainRootDir string) error {
	chainDBDir := filepath.Join(chainRootDir, DataDir)

	blockdb := filepath.Join(chainDBDir, "blockstore.db")
	state := filepath.Join(chainDBDir, "state.db")
	wal := filepath.Join(chainDBDir, "cs.wal")
	evidence := filepath.Join(chainDBDir, "evidence.db")
	txIndex := filepath.Join(chainDBDir, "tx_index.db")

	if cmtos.FileExists(blockdb) {
		if err := os.RemoveAll(blockdb); err == nil {
			fmt.Println("Removed all blockstore.db", "dir", blockdb)
		} else {
			fmt.Println("error removing all blockstore.db", "dir", blockdb, "err", err)
		}
	}

	if cmtos.FileExists(state) {
		if err := os.RemoveAll(state); err == nil {
			fmt.Println("Removed all state.db", "dir", state)
		} else {
			fmt.Println("error removing all state.db", "dir", state, "err", err)
		}
	}

	if cmtos.FileExists(wal) {
		if err := os.RemoveAll(wal); err == nil {
			fmt.Println("Removed all cs.wal", "dir", wal)
		} else {
			fmt.Println("error removing all cs.wal", "dir", wal, "err", err)
		}
	}

	if cmtos.FileExists(evidence) {
		if err := os.RemoveAll(evidence); err == nil {
			fmt.Println("Removed all evidence.db", "dir", evidence)
		} else {
			fmt.Println("error removing all evidence.db", "dir", evidence, "err", err)
		}
	}

	if cmtos.FileExists(txIndex) {
		if err := os.RemoveAll(txIndex); err == nil {
			fmt.Println("Removed all tx_index.db", "dir", txIndex)
		} else {
			fmt.Println("error removing all tx_index.db", "dir", txIndex, "err", err)
		}
	}

	if err := cmtos.EnsureDir(chainDBDir, 0700); err != nil {
		fmt.Println("unable to recreate chainDBDir", "err", err)
	}
	return nil
}
