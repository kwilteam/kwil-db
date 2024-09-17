package config

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

// ResetAll removes all data.
func ResetAll(rootDir string) error {
	// Remove CometBFT's stuff first.
	chainRoot := ABCIDir(rootDir)

	// Address book e.g. <root>/abci/config/addrbook.json
	addrBookFile := cometbft.AddrBookPath(chainRoot)
	if err := os.Remove(addrBookFile); err == nil {
		fmt.Println("Removed existing address book", "file", addrBookFile)
	} else if !os.IsNotExist(err) {
		fmt.Println("Error removing address book", "file", addrBookFile, "err", err)
	}

	// all ABCI data
	abciDir := ABCIDir(rootDir)
	if err := os.RemoveAll(abciDir); err == nil {
		fmt.Println("Removed all ABCI data at directory", abciDir)
	} else {
		fmt.Println("Error removing all ABCI data at directory", abciDir, "err", err)
	}

	// kwild application data
	sigDir := SigningDir(rootDir)
	if err := os.RemoveAll(sigDir); err == nil {
		fmt.Println("Removed all signing at directory", sigDir)
	} else {
		fmt.Println("Error removing all signing at directory", sigDir, "err", err)
	}

	rcvdSnaps := ReceivedSnapshotsDir(rootDir)
	if err := os.RemoveAll(rcvdSnaps); err == nil {
		fmt.Println("Removed all rcvdSnaps at directory", rcvdSnaps)
	} else {
		fmt.Println("Error removing all rcvdSnaps at directory", rcvdSnaps, "err", err)
	}

	migrationDir := MigrationDir(rootDir)
	if err := os.RemoveAll(migrationDir); err == nil {
		fmt.Println("Removed all migrations at directory", migrationDir)
	} else {
		fmt.Println("Error removing all migrations at directory", migrationDir, "err", err)
	}

	// The user-configurable paths

	// TODO: support postgres database drop or schema drops

	snapshotDir := LocalSnapshotsDir(rootDir)

	if err := os.RemoveAll(snapshotDir); err == nil {
		fmt.Println("Removed all snapshots at directory", snapshotDir)
	} else {
		fmt.Println("Error removing all snapshots at directory", snapshotDir, "err", err)
	}

	return nil
}
