package abci

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/snapshots"
)

// resetAll removes address book files plus all data
func ResetAll(cfg *config.KwildConfig) error {
	RemoveAddrBook(cfg.ChainCfg.P2P.AddrBookFile())

	dbDir := cfg.ChainCfg.DBDir()
	if err := os.RemoveAll(dbDir); err == nil {
		fmt.Println("Removed all blockchain history", "dir", dbDir)
	} else {
		fmt.Println("Error removing all blockchain history", "dir", dbDir, "err", err)
	}

	if err := cmtos.EnsureDir(dbDir, 0700); err != nil {
		fmt.Println("Error recreating dbDir", "dir", dbDir, "err", err)
	}

	infoDir := filepath.Join(cfg.ChainCfg.RootDir, "info")
	if err := os.RemoveAll(infoDir); err == nil {
		fmt.Println("Removed all info", "dir", infoDir)
	} else {
		fmt.Println("Error removing all info", "dir", infoDir, "err", err)
	}

	appDir := filepath.Join(cfg.RootDir, "application")
	if err := os.RemoveAll(appDir); err == nil {
		fmt.Println("Removed all application", "dir", appDir)
	} else {
		fmt.Println("Error removing all application", "dir", appDir, "err", err)
	}

	sigDir := filepath.Join(cfg.RootDir, "signing")
	if err := os.RemoveAll(sigDir); err == nil {
		fmt.Println("Removed all signing", "dir", sigDir)
	} else {
		fmt.Println("Error removing all signing", "dir", sigDir, "err", err)
	}

	if err := os.RemoveAll(cfg.AppCfg.SqliteFilePath); err == nil {
		fmt.Println("Removed all sqlite files", "dir", cfg.AppCfg.SqliteFilePath)
	} else {
		fmt.Println("Error removing all sqlite files", "dir", cfg.AppCfg.SqliteFilePath, "err", err)
	}

	rcvdSnaps := filepath.Join(cfg.RootDir, "rcvdSnaps")
	if err := os.RemoveAll(rcvdSnaps); err == nil {
		fmt.Println("Removed all rcvdSnaps", "dir", rcvdSnaps)
	} else {
		fmt.Println("Error removing all rcvdSnaps", "dir", rcvdSnaps, "err", err)
	}

	if err := os.RemoveAll(cfg.AppCfg.SnapshotConfig.SnapshotDir); err == nil {
		fmt.Println("Removed all snapshots", "dir", cfg.AppCfg.SnapshotConfig.SnapshotDir)
	} else {
		fmt.Println("Error removing all snapshots", "dir", cfg.AppCfg.SnapshotConfig.SnapshotDir, "err", err)
	}

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

func RemoveAddrBook(addrBookFile string) {
	if err := os.Remove(addrBookFile); err == nil {
		fmt.Println("Removed existing address book", "file", addrBookFile)
	} else if !os.IsNotExist(err) {
		fmt.Println("Error removing address book", "file", addrBookFile, "err", err)
	}
}

// cometAddrFromPubKey computes the cometBFT address from an ed25519 public key.
// This helper is required to support the application's BeginBlock method where
// the RequestBeginBlock request type includes the addresses of any byzantine
// validators rather than their public keys, as with all of the other methods.
func cometAddrFromPubKey(pubkey []byte) string {
	publicKey := ed25519.PubKey(pubkey)
	return publicKey.Address().String()
}

func convertABCISnapshots(req *abciTypes.Snapshot) *snapshots.Snapshot {
	var metadata snapshots.SnapshotMetadata
	err := json.Unmarshal(req.Metadata, &metadata)
	if err != nil {
		return nil
	}

	snapshot := &snapshots.Snapshot{
		Height:     req.Height,
		Format:     req.Format,
		ChunkCount: req.Chunks,
		Hash:       req.Hash,
		Metadata:   metadata,
	}
	return snapshot
}

func convertToABCISnapshot(snapshot *snapshots.Snapshot) (*abciTypes.Snapshot, error) {
	metadata, err := json.Marshal(snapshot.Metadata)
	if err != nil {
		return nil, err
	}

	return &abciTypes.Snapshot{
		Height:   snapshot.Height,
		Format:   snapshot.Format,
		Chunks:   snapshot.ChunkCount,
		Hash:     snapshot.Hash,
		Metadata: metadata,
	}, nil
}

func abciStatus(status snapshots.Status) abciTypes.ResponseApplySnapshotChunk_Result {
	switch status {
	case snapshots.ACCEPT:
		return abciTypes.ResponseApplySnapshotChunk_ACCEPT
	case snapshots.REJECT:
		return abciTypes.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
	case snapshots.RETRY:
		return abciTypes.ResponseApplySnapshotChunk_RETRY
	default:
		return abciTypes.ResponseApplySnapshotChunk_UNKNOWN
	}
}
