package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kwilteam/kwil-db/common/chain"
	commonCfg "github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types"
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
}

func PrepareForMigration(ctx context.Context, kwildCfg *commonCfg.KwildConfig, genesisCfg *chain.GenesisConfig) (*commonCfg.KwildConfig, *chain.GenesisConfig, error) {
	if kwildCfg.MigrationConfig.MigrateFrom == "" {
		return nil, nil, errors.New("migrate_from is mandatory for migration")
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
		snapshotFileName: filepath.Join(kwildCfg.RootDir, "snapshot.sql.gz"),
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
func (m *migrationClient) pollForGenesisState(ctx context.Context) error {
	// Poll for the genesis state from the old chain
	delay := defaultPollFrequency
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	fmt.Println("Polling for genesis state from the old chain")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Get the genesis state from the old chain
			metadata, err := m.clt.GenesisState(ctx)
			if err != nil {
				continue
			}

			// Check if the genesis state is ready
			if metadata.MigrationState.Status == types.NoActiveMigration || metadata.MigrationState.Status == types.MigrationNotStarted {
				continue
			}

			// Genesis state should ready
			if metadata.SnapshotMetadata == nil || metadata.GenesisConfig == nil {
				continue
			}

			fmt.Println("Genesis state is available for download")
			if err = m.downloadGenesisState(ctx, metadata); err != nil {
				return err
			}

			return nil
		}
	}
}

func (m *migrationClient) downloadGenesisState(ctx context.Context, metadata *types.MigrationMetadata) error {
	var genCfg chain.GenesisConfig
	if err := json.Unmarshal(metadata.GenesisConfig, &genCfg); err != nil {
		return err
	}

	// Save the genesis state
	var snapshotMetadata statesync.Snapshot
	if err := json.Unmarshal(metadata.SnapshotMetadata, &snapshotMetadata); err != nil {
		return err // should we continue polling?
	}

	// create snapshot file
	genesisStateFile, err := os.Create(m.snapshotFileName)
	if err != nil {
		return err
	}

	// retrieve all the snapshot chunks
	for i := uint32(0); i < snapshotMetadata.ChunkCount; i++ {
		chunk, err := m.clt.GenesisSnapshotChunk(ctx, snapshotMetadata.Height, i)
		if err != nil {
			return err
		}
		n, err := genesisStateFile.Write(chunk)
		if err != nil {
			return err
		}
		if n != len(chunk) {
			return err
		}
	}

	// Update the genesis config
	m.genesisCfg.DataAppHash = genCfg.DataAppHash
	m.genesisCfg.Validators = genCfg.Validators
	m.genesisCfg.ConsensusParams.Migration = genCfg.ConsensusParams.Migration

	// Update the kwild config
	m.kwildCfg.AppConfig.GenesisState = m.snapshotFileName
	// Listener extension
	extensions := m.kwildCfg.AppConfig.Extensions
	extensions["migrations"] = map[string]string{
		"listen_address": m.listenAddress,
		"start_height":   fmt.Sprintf("%d", metadata.MigrationState.StartHeight),
		"end_height":     fmt.Sprintf("%d", metadata.MigrationState.EndHeight),
	}
	m.kwildCfg.AppConfig.Extensions = extensions

	fmt.Println("Genesis state downloaded successfully", m.snapshotFileName)
	return nil
}
