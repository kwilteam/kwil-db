package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/internal/migrations"
	"github.com/spf13/cobra"
)

var (
	genesisFileName  = "genesis.json"
	snapshotFileName = "snapshot.sql.gz"

	genesisStateLong = "Download the genesis state corresponding to the on-going migration. The genesis state includes the genesis config file(`genesis.json`), genesis snapshot(`snapshot.sql.gz`), and the migration info such as the start and end heights, if migration is active etc. The genesis state is saved in the root directory specified by the `--root-dir` flag. If there is no migration in progress, the command will log `no active migration` and no files will be downloaded in this scenario."
)

func genesisStateCmd() *cobra.Command {
	var rootDir string
	cmd := &cobra.Command{
		Use:   "genesis-state",
		Short: "Download the genesis state corresponding to the on-going migration.",
		Long:  genesisStateLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// Request for the genesis state {genesis file data, snapshot metadata, migration state: active, start, endHeight}
			inMigration, state, err := clt.GenesisState(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if !inMigration {
				return display.PrintCmd(cmd, &MigrationState{InMigration: false})
			}

			// ensure the root directory exists
			expandedDir, err := common.ExpandPath(rootDir)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if err = os.MkdirAll(expandedDir, 0755); err != nil {
				return display.PrintErr(cmd, err)
			}

			var metadata migrations.MigrationMetadata
			if err := metadata.UnmarshalBinary(state); err != nil {
				return display.PrintErr(cmd, err)
			}

			// Print the genesis state to the genesis.json file
			genesisFile := filepath.Join(expandedDir, genesisFileName)
			if metadata.GenesisConfig == nil {
				return display.PrintCmd(cmd, &MigrationState{
					InMigration: metadata.InMigration,
					StartHeight: metadata.StartHeight,
					EndHeight:   metadata.EndHeight,
					GenesisFile: "",
					Snapshot:    "",
				})
			}

			err = metadata.GenesisConfig.SaveAs(genesisFile)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// Retrieve snapshot chunks and save them to the snapshot file
			if metadata.GenesisSnapshot == nil {
				return display.PrintCmd(cmd, &MigrationState{
					InMigration: metadata.InMigration,
					StartHeight: metadata.StartHeight,
					EndHeight:   metadata.EndHeight,
					GenesisFile: genesisFile,
					Snapshot:    "",
				})
			}

			// create snapshot file
			snapshotFile := filepath.Join(expandedDir, snapshotFileName)
			snapshot, err := os.Create(snapshotFile)
			if err != nil {
				return display.PrintErr(cmd, err)
			}
			defer snapshot.Close()

			snapshotHeight := metadata.GenesisSnapshot.Height
			// retrieve all the snapshot chunks
			for i := uint32(0); i < metadata.GenesisSnapshot.ChunkCount; i++ {
				chunk, err := clt.GenesisSnapshotChunk(ctx, snapshotHeight, i)
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
	InMigration bool
	StartHeight int64
	EndHeight   int64
	GenesisFile string
	Snapshot    string
}

func (m *MigrationState) MarshalText() ([]byte, error) {
	if !m.InMigration {
		return []byte("No active migration in progress."), nil
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
	return json.Marshal(m)
}
