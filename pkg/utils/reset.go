package utils

import (
	"fmt"
	"os"
	"path/filepath"

	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/privval"
)

// resetAll removes address book files plus all data, and resets the privValdiator data.
func ResetAll(dbDir, addrBookFile, privValKeyFile, privValStateFile string) error {
	RemoveAddrBook(addrBookFile)

	if err := os.RemoveAll(dbDir); err == nil {
		fmt.Println("Removed all blockchain history", "dir", dbDir)
	} else {
		fmt.Println("Error removing all blockchain history", "dir", dbDir, "err", err)
	}

	if err := cmtos.EnsureDir(dbDir, 0700); err != nil {
		fmt.Println("Error recreating dbDir", "dir", dbDir, "err", err)
	}

	// recreate the dbDir since the privVal state needs to live there
	ResetFilePV(privValKeyFile, privValStateFile)
	return nil
}

// resetState removes address book files plus all databases.
func ResetState(dbDir string) error {
	blockdb := filepath.Join(dbDir, "blockstore.db")
	state := filepath.Join(dbDir, "state.db")
	wal := filepath.Join(dbDir, "cs.wal")
	evidence := filepath.Join(dbDir, "evidence.db")
	txIndex := filepath.Join(dbDir, "tx_index.db")

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

	if err := cmtos.EnsureDir(dbDir, 0700); err != nil {
		fmt.Println("unable to recreate dbDir", "err", err)
	}
	return nil
}

func ResetFilePV(privValKeyFile, privValStateFile string) {
	if _, err := os.Stat(privValKeyFile); err == nil {
		pv := privval.LoadFilePVEmptyState(privValKeyFile, privValStateFile)
		pv.Reset()
		fmt.Println("Reset private validator file to genesis state", "keyFile", privValKeyFile, "stateFile", privValStateFile)
	} else {
		pv := privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		fmt.Println("Generated private validator file", "keyFile", privValKeyFile, "stateFile", privValStateFile)
	}
}

func RemoveAddrBook(addrBookFile string) {
	if err := os.Remove(addrBookFile); err == nil {
		fmt.Println("Removed existing address book", "file", addrBookFile)
	} else if !os.IsNotExist(err) {
		fmt.Println("Error removing address book", "file", addrBookFile, "err", err)
	}
}
