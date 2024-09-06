package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

func ResetChainState(rootDir string) error {
	chainRoot := filepath.Join(rootDir, ABCIDirName)
	return cometbft.ResetState(chainRoot)
}

// ResetAll removes all data.
func ResetAll(rootDir, snapshotDir string) error {
	// Remove CometBFT's stuff first.
	chainRoot := filepath.Join(rootDir, ABCIDirName)

	// Address book e.g. <root>/abci/config/addrbook.json
	addrBookFile := cometbft.AddrBookPath(chainRoot)
	if err := os.Remove(addrBookFile); err == nil {
		fmt.Println("Removed existing address book", "file", addrBookFile)
	} else if !os.IsNotExist(err) {
		fmt.Println("Error removing address book", "file", addrBookFile, "err", err)
	}

	// Blockchain data files. e.g. <root>/abci/data
	dbDir := filepath.Join(chainRoot, cometbft.DataDir)
	if err := os.RemoveAll(dbDir); err == nil {
		fmt.Println("Removed all blockchain history", "dir", dbDir)
	} else {
		fmt.Println("Error removing all blockchain history", "dir", dbDir, "err", err)
	}
	// wasn't that ResetState?
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		fmt.Println("Error recreating dbDir", "dir", dbDir, "err", err)
	}

	// kwild application data

	infoDir := filepath.Join(chainRoot, ABCIInfoSubDirName)
	if err := os.RemoveAll(infoDir); err == nil {
		fmt.Println("Removed all info", "dir", infoDir)
	} else {
		fmt.Println("Error removing all info", "dir", infoDir, "err", err)
	}

	sigDir := filepath.Join(rootDir, SigningDirName)
	if err := os.RemoveAll(sigDir); err == nil {
		fmt.Println("Removed all signing", "dir", sigDir)
	} else {
		fmt.Println("Error removing all signing", "dir", sigDir, "err", err)
	}

	rcvdSnaps := filepath.Join(rootDir, ReceivedSnapsDirName)
	if err := os.RemoveAll(rcvdSnaps); err == nil {
		fmt.Println("Removed all rcvdSnaps", "dir", rcvdSnaps)
	} else {
		fmt.Println("Error removing all rcvdSnaps", "dir", rcvdSnaps, "err", err)
	}

	migrationDir := filepath.Join(rootDir, MigrationsDirName)
	if err := os.RemoveAll(migrationDir); err == nil {
		fmt.Println("Removed all migrations", "dir", migrationDir)
	} else {
		fmt.Println("Error removing all migrations", "dir", migrationDir, "err", err)
	}

	// The user-configurable paths

	// TODO: support postgres database drop or schema drops

	snapshotDir, err := config.CleanPath(snapshotDir, rootDir)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(snapshotDir); err == nil {
		fmt.Println("Removed all snapshots", "dir", snapshotDir)
	} else {
		fmt.Println("Error removing all snapshots", "dir", snapshotDir, "err", err)
	}

	return nil
}
