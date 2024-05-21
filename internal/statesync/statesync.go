package statesync

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/voting"

	cometClient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
)

const (
	ABCISnapshotQueryPath        = "/snapshot/height"
	ABCILatestSnapshotHeightPath = "/snapshot/latest"
)

// StateSyncer is responsible for initializing the database state from the
// snapshots received from the network. It validates the snapshots against
// the trusted snapshot providers.
// The snapshot used to initialize are discarded after the database is restored.
// Snapshot store if enabled, only stores snapshots produced by the node itself.

type StateSyncer struct {
	// statesync configuration
	dbConfig *DBConfig

	db sql.ReadTxMaker

	// directory to store snapshots and chunks
	// same as the snapshot store directory
	// as we allow to reuse the received snapshots
	snapshotsDir string

	// trusted snapshot providers for verification - cometbft rfc servers
	snapshotProviders []*rpchttp.HTTP

	// State syncer state
	snapshot   *Snapshot
	chunks     map[uint32]bool // Chunks received till now
	rcvdChunks uint32          // Number of chunks received till now

	// Logger
	log log.Logger
}

// NewStateSyncer will initialize the state syncer that enables the node to
// receive and validate snapshots from the network and initialize the database state.
// It takes the database configuration, snapshot directory, and the trusted snapshot providers.
// Trusted snapshot providers are special nodes in the network trusted by the nodes and
// have snapshot creation enabled. These nodes are responsible for creating and validating snapshots.
func NewStateSyncer(ctx context.Context, cfg *DBConfig, snapshotDir string, providers []string, db sql.ReadTxMaker, logger log.Logger) *StateSyncer {

	ss := &StateSyncer{
		dbConfig:     cfg,
		db:           db,
		snapshotsDir: snapshotDir,
		log:          logger,
	}

	for _, s := range providers {
		clt, err := ChainRPCClient(s)
		if err != nil {
			logger.Error("Failed to create rpc client", log.String("server", s), log.Error(err))
			return nil
		}
		ss.snapshotProviders = append(ss.snapshotProviders, clt)
	}

	// Ensure that the snapshot directory exists and is empty
	if err := os.RemoveAll(snapshotDir); err != nil {
		logger.Error("Failed to delete snapshot directory", log.String("dir", snapshotDir), log.Error(err))
		return nil
	}
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		logger.Error("Failed to create snapshot directory", log.String("dir", snapshotDir), log.Error(err))
		return nil
	}

	return ss
}

// OfferSnapshot checks if the snapshot is valid and kicks off the state sync process
// accepted snapshot is stored on disk for later processing
func (ss *StateSyncer) OfferSnapshot(ctx context.Context, snapshot *Snapshot) error {
	ss.log.Info("Offering snapshot", log.Int("height", int64(snapshot.Height)), log.Uint("format", snapshot.Format), log.String("App Hash", fmt.Sprintf("%x", snapshot.SnapshotHash)))

	// Check if we are already in the middle of a snapshot
	if ss.snapshot != nil {
		return ErrStateSyncInProgress
	}

	// Validate the snapshot
	err := ss.validateSnapshot(ctx, *snapshot)
	if err != nil {
		return err
	}

	ss.snapshot = snapshot
	ss.chunks = make(map[uint32]bool, snapshot.ChunkCount)
	ss.rcvdChunks = 0
	return nil
}

// ApplySnapshotChunk accepts a chunk and stores it on disk for later processing if valid
// If all chunks are received, it starts the process of restoring the database
func (ss *StateSyncer) ApplySnapshotChunk(ctx context.Context, chunk []byte, index uint32) (bool, error) {
	if ss.snapshot == nil {
		return false, ErrStateSyncNotInProgress
	}

	// Check if the chunk has already been applied
	if ss.chunks[index] {
		return false, nil
	}

	// Check if the chunk index is valid
	if index >= ss.snapshot.ChunkCount {
		ss.log.Error("Invalid chunk index", log.Uint("index", index), log.Uint("chunk-count", ss.snapshot.ChunkCount))
		return false, ErrRejectSnapshotChunk
	}

	// Validate the chunk hash
	chunkHash := sha256.Sum256(chunk)
	if chunkHash != ss.snapshot.ChunkHashes[index] {
		return false, ErrRefetchSnapshotChunk
	}

	// store the chunk on disk
	chunkFileName := filepath.Join(ss.snapshotsDir, fmt.Sprintf("chunk-%d.sql.gz", index))
	err := os.WriteFile(chunkFileName, chunk, 0755)
	if err != nil {
		os.Remove(chunkFileName)
		return false, errors.Join(err, ErrRetrySnapshotChunk)
	}

	ss.log.Info("Applied snapshot chunk", log.Uint("height", ss.snapshot.Height), log.Uint("index", index))

	ss.chunks[index] = true
	ss.rcvdChunks++

	// Kick off the process of restoring the database if all chunks are received
	if ss.rcvdChunks == ss.snapshot.ChunkCount {
		ss.log.Info("All chunks received - Starting DB restore process")

		// Ensure that the DB is empty before applying the snapshot
		initialized, err := isDbInitialized(ctx, ss.db)
		if err != nil {
			ss.resetStateSync()
			return false, errors.Join(err, ErrRejectSnapshot)
		}

		if initialized {
			ss.resetStateSync()
			// Statesync is not allowed on an initialized DB
			return false, errors.Join(ErrAbortSnapshotChunk, errors.New("postgres DB state is not empty, please reset the DB state before applying the snapshot"))
		}

		// Restore the DB from the chunks
		streamer := NewStreamer(ss.snapshot.ChunkCount, ss.snapshotsDir, ss.log)
		defer streamer.Close()
		reader, err := gzip.NewReader(streamer)
		if err != nil {
			ss.resetStateSync()
			return false, errors.Join(err, ErrRejectSnapshot)
		}
		defer reader.Close()

		err = RestoreDB(ctx, reader, ss.dbConfig.DBName, ss.dbConfig.DBUser,
			ss.dbConfig.DBPass, ss.dbConfig.DBHost, ss.dbConfig.DBPort,
			ss.snapshot.SnapshotHash, ss.log)
		if err != nil {
			ss.resetStateSync()
			return false, errors.Join(err, ErrRejectSnapshot)
		}
		ss.log.Info("DB restored")

		ss.chunks = nil
		ss.rcvdChunks = 0
		ss.snapshot = nil
		return true, nil
	}

	return false, nil
}

// RestoreDB restores the database from the logical sql dump using psql command
// It also validates the snapshot hash, before restoring the database
func RestoreDB(ctx context.Context, snapshot io.Reader,
	dbName, dbUser, dbPass, dbHost, dbPort string,
	snapshotHash []byte, logger log.Logger) error {
	// unzip and stream the sql dump to psql
	cmd := exec.CommandContext(ctx,
		"psql",
		"--username", dbUser,
		"--host", dbHost,
		"--port", dbPort,
		"--dbname", dbName,
		"--no-password",
	)
	if dbPass != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
	}

	// cmd.Stdout = &stderr
	stdinPipe, err := cmd.StdinPipe() // stdin for psql command
	if err != nil {
		return err
	}
	defer stdinPipe.Close()

	logger.Info("Restore DB: ", log.String("command", cmd.String()))

	if err := cmd.Start(); err != nil {
		return err
	}

	// decompress the chunk streams and stream the sql dump to psql stdinPipe
	if err := decompressAndValidateSnapshotHash(stdinPipe, snapshot, snapshotHash); err != nil {
		return err
	}

	stdinPipe.Close() // signifies the end of the input stream to the psql command

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// decompressAndValidateChunkStreams decompresses the chunk streams and validates the snapshot hash
func decompressAndValidateSnapshotHash(output io.Writer, reader io.Reader, snapshotHash []byte) error {
	hasher := sha256.New()
	_, err := io.Copy(io.MultiWriter(output, hasher), reader)
	if err != nil {
		return fmt.Errorf("failed to decompress chunk streams: %w", err)
	}
	hash := hasher.Sum(nil)

	// Validate the hash of the decompressed chunks
	if !bytes.Equal(hash, snapshotHash) {
		return fmt.Errorf("invalid snapshot hash %x, expected %x", hash, snapshotHash)
	}
	return nil
}

// validateSnapshot validates the snapshot against the trusted snapshot providers
func (ss *StateSyncer) validateSnapshot(ctx context.Context, snapshot Snapshot) error {
	// Check if the snapshot's contents are valid
	if snapshot.Format != DefaultSnapshotFormat {
		return ErrUnsupportedSnapshotFormat
	}

	if snapshot.Height <= 0 || snapshot.ChunkCount <= 0 ||
		snapshot.ChunkCount != uint32(len(snapshot.ChunkHashes)) {
		return ErrInvalidSnapshot
	}

	// Query the snapshot providers to check if the snapshot is valid
	height := fmt.Sprintf("%d", snapshot.Height)
	verified := false
	for _, clt := range ss.snapshotProviders {
		res, err := clt.ABCIQuery(ctx, ABCISnapshotQueryPath, []byte(height))
		if err != nil {
			ss.log.Info("Failed to query snapshot", log.Error(err)) // failover to next provider
			continue
		}

		if len(res.Response.Value) > 0 {
			var snap Snapshot
			err = json.Unmarshal(res.Response.Value, &snap)
			if err != nil {
				ss.log.Error("Failed to unmarshal snapshot", log.Error(err))
				continue
			}

			if snap.Height != snapshot.Height || snap.SnapshotSize != snapshot.SnapshotSize ||
				snap.ChunkCount != snapshot.ChunkCount || !bytes.Equal(snap.SnapshotHash, snapshot.SnapshotHash) {
				ss.log.Error("Invalid snapshot", log.Uint("height", snapshot.Height), log.Any("Expected ", snap), log.Any("Actual", snapshot))
				break
			}

			verified = true
			break
		}
	}

	if !verified {
		return ErrInvalidSnapshot
	}

	return nil
}

func (ss *StateSyncer) resetStateSync() {
	ss.snapshot = nil
	ss.chunks = nil
	ss.rcvdChunks = 0

	os.RemoveAll(ss.snapshotsDir)
	os.MkdirAll(ss.snapshotsDir, 0755)
}

// rpcClient sets up a new RPC client
func ChainRPCClient(server string) (*rpchttp.HTTP, error) {
	if !strings.Contains(server, "://") {
		server = "http://" + server
	}
	c, err := rpchttp.New(server)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetLatestSnapshotHeight queries the trusted snapshot providers to get the latest snapshot height.
func GetLatestSnapshotInfo(ctx context.Context, client cometClient.ABCIClient) (*Snapshot, error) {
	res, err := client.ABCIQuery(ctx, ABCILatestSnapshotHeightPath, nil)
	if err != nil {
		return nil, err
	}

	if len(res.Response.Value) == 0 {
		return nil, errors.New("no snapshot found")
	}

	var snap Snapshot
	err = json.Unmarshal(res.Response.Value, &snap)
	if err != nil {
		return nil, err
	}

	return &snap, nil
}

func isDbInitialized(ctx context.Context, db sql.ReadTxMaker) (bool, error) {
	tx, err := db.BeginReadTx(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	vals, err := voting.GetValidators(ctx, tx)
	if err != nil {
		return false, err
	}

	return len(vals) > 0, nil
}
