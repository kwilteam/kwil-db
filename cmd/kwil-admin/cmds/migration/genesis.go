package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/migrations"
	"github.com/kwilteam/kwil-db/internal/statesync"
)

var (
	genesisFileName  = "genesis.json"
	snapshotFileName = "snapshot.sql.gz"

	genesisStateLong = "Download the genesis state for the new network from a trusted node on the source network. The genesis state includes the genesis config file (genesis.json), genesis snapshot (snapshot.sql.gz), and the migration info such as the start and end heights. The genesis state is saved in the root directory specified by the `--out-dir` flag. If there is no approved migration or if the migration has not started yet, the command will return a message indicating that there is no genesis state to download."

	genesisStateExample = `# Download the genesis state to the default output directory (~/.genesis-state)
kwil-admin migrate genesis-state

# Download the genesis state to a custom root directory
kwil-admin migrate genesis-state --out-dir /path/to/out/dir`
)

func genesisStateCmd() *cobra.Command {
	var rootDir string
	cmd := &cobra.Command{
		Use:     "genesis-state",
		Short:   "Download the genesis state corresponding to the ongoing migration.",
		Long:    genesisStateLong,
		Example: genesisStateExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// Request for the genesis state {genesis file data, snapshot metadata, migration state: active, start, endHeight}
			metadata, err := clt.GenesisState(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// this check should change in every version:
			// For backwards compatibility, we should be able to unmarshal structs from previous versions.
			// Since v0.9 is our first time supporting migration, we only need to check for v0.9.
			if metadata.Version != migrations.MigrationVersion {
				return display.PrintErr(cmd, fmt.Errorf("genesis state download is incompatible. Received version: %d, supported versions: [%d]", metadata.Version, migrations.MigrationVersion))
			}

			// If there is no active migration or if the migration has not started yet, return the migration state
			// indicating that there is no genesis state to download.
			if metadata.MigrationState.Status == types.NoActiveMigration ||
				metadata.MigrationState.Status == types.ActivationPeriod ||
				metadata.GenesisInfo == nil || metadata.SnapshotMetadata == nil {
				return display.PrintCmd(cmd, &MigrationState{
					Info: metadata.MigrationState,
				})
			}

			// ensure the root directory exists
			expandedDir, err := common.ExpandPath(rootDir)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if err = os.MkdirAll(expandedDir, 0755); err != nil {
				return display.PrintErr(cmd, err)
			}

			// retrieve the snapshot metadata
			var snapshotMetadata statesync.Snapshot
			if err = json.Unmarshal(metadata.SnapshotMetadata, &snapshotMetadata); err != nil {
				return display.PrintErr(cmd, err)
			}

			genCfg := chain.GenesisConfig{
				DataAppHash:   metadata.GenesisInfo.AppHash,
				InitialHeight: metadata.MigrationState.StartHeight,
				ConsensusParams: &chain.ConsensusParams{
					BaseConsensusParams: chain.BaseConsensusParams{
						Migration: chain.MigrationParams{
							StartHeight: metadata.MigrationState.StartHeight,
							EndHeight:   metadata.MigrationState.EndHeight,
						},
					},
				},
			}

			for _, nv := range metadata.GenesisInfo.Validators {
				genCfg.Validators = append(genCfg.Validators, &chain.GenesisValidator{
					Name:   nv.Name,
					PubKey: nv.PubKey,
					Power:  nv.Power,
				})
			}

			// Print the genesis state to the genesis.json file
			genesisFile := filepath.Join(expandedDir, genesisFileName)
			err = genCfg.SaveAs(genesisFile)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// create snapshot file
			snapshotFile := filepath.Join(expandedDir, snapshotFileName)
			snapshot, err := os.Create(snapshotFile)
			if err != nil {
				return display.PrintErr(cmd, err)
			}
			defer snapshot.Close()

			// retrieve all the snapshot chunks
			for i := uint32(0); i < snapshotMetadata.ChunkCount; i++ {
				chunk, err := clt.GenesisSnapshotChunk(ctx, snapshotMetadata.Height, i)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
				n, err := snapshot.Write(chunk)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
				if n != len(chunk) {
					return display.PrintErr(cmd, fmt.Errorf("failed to write snapshot chunk to file"))
				}
			}

			// Print the migration state
			return display.PrintCmd(cmd, &MigrationState{
				Info:        metadata.MigrationState,
				GenesisFile: genesisFile,
				Snapshot:    snapshotFile,
			})

		},
	}

	common.BindRPCFlags(cmd)
	cmd.Flags().StringVarP(&rootDir, "out-dir", "o", "~/.genesis-state", "The target directory for downloading the genesis state files.")
	return cmd
}

type MigrationState struct {
	Info        types.MigrationState `json:"state"`
	GenesisFile string               `json:"genesis_file"`
	Snapshot    string               `json:"snapshot"`
}

func (m *MigrationState) MarshalText() ([]byte, error) {
	if m.Info.Status == types.NoActiveMigration {
		return []byte("No active migration found."), nil
	}

	if m.Info.Status == types.GenesisMigration {
		return []byte("Genesis migration in progress. No genesis state to download."), nil
	}

	if m.Info.Status == types.ActivationPeriod {
		return []byte(fmt.Sprintf("No genesis state to download yet. Migration is set to start at block height: %d", m.Info.StartHeight)), nil
	}

	if m.GenesisFile == "" {
		return []byte("No genesis.json file data found."), nil
	}

	if m.Snapshot == "" {
		return []byte("No snapshot.sql.gz file data found."), nil
	}

	return []byte(fmt.Sprintf("Migration State:\n"+
		"\tStart Height: %d\n"+
		"\tEnd Height: %d\n"+
		"\tGenesis File: %s\n"+
		"\tSnapshot File: %s\n",
		m.Info.StartHeight, m.Info.EndHeight, m.GenesisFile, m.Snapshot)), nil
}

func (m *MigrationState) MarshalJSON() ([]byte, error) {
	type alias MigrationState
	return json.Marshal((*alias)(m)) // slice off methods to avoid recursive call
}
