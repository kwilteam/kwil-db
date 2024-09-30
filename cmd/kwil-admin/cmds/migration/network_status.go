package migration

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

func networkStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "network-status",
		Short: "Get the migration status of the network.",
		Example: `# Get the migration status of the network.
		kwil-admin migrate network-status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			status, err := clt.MigrationStatus(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if status.Status == types.UnknownMigrationStatus {
				return display.PrintErr(cmd, fmt.Errorf("migration status unknown"))
			}

			return display.PrintCmd(cmd, &migrationStatus{
				Status:        status.Status,
				StartHeight:   status.StartHeight,
				EndHeight:     status.EndHeight,
				CurrentHeight: status.CurrentBlock,
			})
		},
	}
}

type migrationStatus struct {
	Status        types.MigrationStatus `json:"status"`
	StartHeight   int64                 `json:"start_height"`
	EndHeight     int64                 `json:"end_height"`
	CurrentHeight int64                 `json:"current_height"`
}

func (m *migrationStatus) MarshalJSON() ([]byte, error) {
	type alias migrationStatus
	return json.Marshal((*alias)(m)) // slice off methods to avoid recursive call
}

func (m *migrationStatus) MarshalText() ([]byte, error) {
	if m.Status == types.NoActiveMigration {
		return []byte("No active migration on the network."), nil
	}

	if m.Status == types.GenesisMigration {
		return []byte("Genesis migration in progress."), nil
	}

	return []byte(fmt.Sprintf("Migration Status: %s\n"+
		"StartHeight: %d\n"+
		"EndHeight: %d\n"+
		"CurrentBlock: %d\n",
		m.Status.String(), m.StartHeight, m.EndHeight, m.CurrentHeight)), nil
}
