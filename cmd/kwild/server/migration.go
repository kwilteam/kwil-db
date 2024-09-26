package server

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
	"path/filepath"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/chain"
	commonCfg "github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/statesync"
)

// The migrationClient type is responsible for:
//   - Polling the old chain to retrieve the genesis state required for migration.
//   - Downloading the genesis snapshot from the old chain and saves it under the
//     root directory of the node.
//   - Updating the genesis configuration such as app hash, validators and migration settings.
//   - Updating the kwild configuration with the snapshot file path and migrations listener extension.

const (
	defaultPollFrequency = time.Second * 30
)

type migrationClient struct {
	// listenAddress is the old chain's listen address to retrieve the genesis state
	listenAddress string

	snapshotFileName string

	clt        *client.Client
	kwildCfg   *commonCfg.KwildConfig
	genesisCfg *chain.GenesisConfig

	logger log.Logger
}

func PrepareForMigration(ctx context.Context, kwildCfg *commonCfg.KwildConfig, genesisCfg *chain.GenesisConfig, logger log.Logger) (*commonCfg.KwildConfig, *chain.GenesisConfig, error) {
	if kwildCfg.MigrationConfig.MigrateFrom == "" {
		return nil, nil, errors.New("migrate_from is mandatory for migration")
	}

	snapshotFileName := config.GenesisStateFileName(kwildCfg.RootDir)
	// check if genesis hash is set in the genesis config
	if genesisCfg.DataAppHash != nil &&
		genesisCfg.ConsensusParams.Migration.StartHeight != 0 &&
		genesisCfg.ConsensusParams.Migration.EndHeight != 0 &&
		validateGenesisState(snapshotFileName, genesisCfg.DataAppHash) {
		// genesis state already downloaded. No need to poll for genesis state
		logger.Info("Genesis state already downloaded", log.String("genesis snapshot", snapshotFileName))
		return kwildCfg, genesisCfg, nil
	}

	// old chain client
	clt, err := client.NewClient(ctx, kwildCfg.MigrationConfig.MigrateFrom, nil)
	if err != nil {
		return nil, nil, err
	}

	// Get the genesis state from the old chain
	m := &migrationClient{
		listenAddress:    kwildCfg.MigrationConfig.MigrateFrom,
		clt:              clt,
		kwildCfg:         kwildCfg,
		genesisCfg:       genesisCfg,
		snapshotFileName: snapshotFileName,
		logger:           logger,
	}

	// poll for the genesis state
	if err = m.pollForGenesisState(ctx); err != nil {
		return nil, nil, err
	}

	return m.kwildCfg, m.genesisCfg, nil
}

// pollForGenesisState polls for the genesis state from the old chain
// and downloads the genesis state to the snapshot file abnd updates the genesis config
// and the kwild config with the configuration required for migration
func (m *migrationClient) pollForGenesisState(ctx context.Context) (err error) {
	// Poll for the genesis state from the old chain
	m.logger.Info("Requesting genesis state from the old chain", log.String("listen_address", m.listenAddress))
	for {
		if err = m.downloadGenesisState(ctx); err == nil {
			return nil
		}
		m.logger.Info("Genesis state not available", log.Error(err), log.Duration("retry after(sec)", defaultPollFrequency))

		// retry after defaultPollFrequency
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultPollFrequency):
		}
	}
}

func (m *migrationClient) downloadGenesisState(ctx context.Context) error {
	// Get the genesis state from the old chain
	metadata, err := m.clt.GenesisState(ctx)
	if err != nil {
		return err
	}

	// Check if the genesis state is ready
	if metadata.MigrationState.Status == types.NoActiveMigration || metadata.MigrationState.Status == types.MigrationNotStarted {
		return fmt.Errorf("status %s", metadata.MigrationState.Status.String())
	}

	// Genesis state should ready
	if metadata.SnapshotMetadata == nil || metadata.GenesisConfig == nil {
		return errors.New("genesis state not available")
	}

	var genCfg chain.GenesisConfig
	if err := json.Unmarshal(metadata.GenesisConfig, &genCfg); err != nil {
		return fmt.Errorf("failed to unmarshal genesis config: %w", err)
	}

	// Save the genesis state
	var snapshotMetadata statesync.Snapshot
	if err := json.Unmarshal(metadata.SnapshotMetadata, &snapshotMetadata); err != nil {
		return fmt.Errorf("failed to unmarshal snapshot metadata: %w", err)
	}

	m.logger.Info("Genesis state available for download")

	// create snapshot file
	genesisStateFile, err := os.Create(m.snapshotFileName)
	if err != nil {
		return fmt.Errorf("failed to create genesis snapshot file: %w", err)
	}

	// retrieve all the snapshot chunks
	for i := uint32(0); i < snapshotMetadata.ChunkCount; i++ {
		chunk, err := m.clt.GenesisSnapshotChunk(ctx, snapshotMetadata.Height, i)
		if err != nil {
			return fmt.Errorf("failed to download genesis snapshot chunk: %d  error: %w", i, err)
		}
		n, err := genesisStateFile.Write(chunk)
		if err != nil {
			return fmt.Errorf("failed to write genesis snapshot chunk: %d  error: %w", i, err)
		}
		if n != len(chunk) {
			return fmt.Errorf("incomplete write to genesis snapshot chunk. expected: %d, written: %d", len(chunk), n)
		}
	}

	// Update the genesis config
	m.genesisCfg.DataAppHash = genCfg.DataAppHash
	m.genesisCfg.Validators = genCfg.Validators
	m.genesisCfg.ConsensusParams.Migration = genCfg.ConsensusParams.Migration
	m.genesisCfg.InitialHeight = metadata.MigrationState.StartHeight

	// persist the genesis config
	if err := m.genesisCfg.SaveAs(filepath.Join(m.kwildCfg.RootDir, cometbft.GenesisJSONName)); err != nil {
		return fmt.Errorf("failed to save genesis config: %w", err)
	}

	// Update the kwild config
	m.kwildCfg.AppConfig.GenesisState = m.snapshotFileName

	// persist the kwild config
	if err := nodecfg.WriteConfigFile(filepath.Join(m.kwildCfg.RootDir, cometbft.ConfigTOMLName), m.kwildCfg); err != nil {
		return fmt.Errorf("failed to save kwild config: %w", err)
	}

	m.logger.Info("Genesis state downloaded successfully", log.String("genesis snapshot", m.snapshotFileName))
	return nil
}

func validateGenesisState(filename string, appHash []byte) bool {
	// check if the genesis state file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	genesisStateFile, err := os.Open(filename)
	if err != nil {
		return false
	}

	// gzip reader and hash reader
	gzipReader, err := gzip.NewReader(genesisStateFile)
	if err != nil {
		failBuild(err, "failed to create gzip reader")
	}
	defer gzipReader.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, gzipReader)
	if err != nil {
		return false
	}

	hash := hasher.Sum(nil)
	return appHash != nil && len(hash) == len(appHash) && bytes.Equal(hash, appHash)
}
