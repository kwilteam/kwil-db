package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/internal/statesync"
	"github.com/spf13/cobra"
)

var (
	genesisFileName  = "genesis.json"
	snapshotFileName = "snapshot.sql.gz"

	genesisStateLong = "Download the genesis state for the new network from a trusted node on the source network. The genesis state includes the genesis config file (genesis.json) , genesis snapshot (snapshot.sql.gz) , and the migration info such as the start and end heights. The genesis state is saved in the root directory specified by the `--root-dir` flag. If there is no approved migration or if the migration has not started yet, the command will return a message indicating that there is no genesis state to download."

	genesisStateExample = `# Download the genesis state to the default root directory (~/.kwild)
kwil-admin migration genesis-state

# Download the genesis state to a custom root directory
kwil-admin migration genesis-state --root-dir /path/to/root/dir.`
)

func genesisStateCmd() *cobra.Command {
	var rootDir string
	cmd := &cobra.Command{
		Use:     "genesis-state",
		Short:   "Download the genesis state corresponding to the on-going migration.",
		Long:    genesisStateLong,
		Example: genesisStateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// Request for the genesis state {genesis file data, snapshot metadata, migration state: active, start, endHeight}
			metadata, err := clt.GenesisState(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if !metadata.InMigration || metadata.GenesisConfig == nil || metadata.SnapshotMetadata == nil {
				return display.PrintCmd(cmd, &MigrationState{InMigration: false, StartHeight: metadata.StartHeight, EndHeight: metadata.EndHeight})
			}

			// ensure the root directory exists
			expandedDir, err := common.ExpandPath(rootDir)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if err = os.MkdirAll(expandedDir, 0755); err != nil {
				return display.PrintErr(cmd, err)
			}

			// retrieve the genesis config
			var genCfg chain.GenesisConfig
			if err = json.Unmarshal(metadata.GenesisConfig, &genCfg); err != nil {
				return display.PrintErr(cmd, err)
			}

			// retrieve the snapshot metadata
			var snapshotMetadata statesync.Snapshot
			if err = json.Unmarshal(metadata.SnapshotMetadata, &snapshotMetadata); err != nil {
				return display.PrintErr(cmd, err)
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
				InMigration: metadata.InMigration,
				StartHeight: metadata.StartHeight,
				EndHeight:   metadata.EndHeight,
				GenesisFile: genesisFile,
				Snapshot:    snapshotFile,
			})

		},
	}

	common.BindRPCFlags(cmd)
	cmd.Flags().StringVar(&rootDir, "root-dir", "~/.kwild", "Root directory for the genesis state files")
	return cmd
}

type MigrationState struct {
	InMigration bool   `json:"in_migration"`
	StartHeight int64  `json:"start_height"`
	EndHeight   int64  `json:"end_height"`
	GenesisFile string `json:"genesis_file"`
	Snapshot    string `json:"snapshot"`
}

func (m *MigrationState) MarshalText() ([]byte, error) {
	if !m.InMigration {
		return []byte(fmt.Sprintf("No genesis state to download yet. Migration is set to start at block height: %d", m.StartHeight)), nil
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
		m.StartHeight, m.EndHeight, m.GenesisFile, m.Snapshot)), nil
}

func (m *MigrationState) MarshalJSON() ([]byte, error) {
	type alias MigrationState
	return json.Marshal((*alias)(m)) // slice off methods to avoid recursive call
}
