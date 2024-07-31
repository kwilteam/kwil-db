package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the pending migration resolutions.",
		Long:  "List all the pending migration resolutions.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			migrations, err := clt.ListMigrations(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &MigrationsList{migrations: migrations})
		},
	}

	return cmd
}

type MigrationsList struct {
	migrations []*types.Migration
}

func (m *MigrationsList) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.migrations)
}

func (m *MigrationsList) MarshalText() ([]byte, error) {
	if len(m.migrations) == 0 {
		return []byte("No migrations found."), nil
	}

	var msg bytes.Buffer
	msg.WriteString("Migrations:\n")

	for _, migration := range m.migrations {
		msg.WriteString(fmt.Sprintf("%s:\n", migration.ID))
		msg.WriteString(fmt.Sprintf("\tactivationHeight: %s\n", migration.ActivationHeight.String()))
		msg.WriteString(fmt.Sprintf("\tmigrationDuration: %s\n", migration.MigrationDuration.String()))
		msg.WriteString(fmt.Sprintf("\tchainID: %s\n", migration.ChainID))
		msg.WriteString(fmt.Sprintf("\ttimestamp: %s\n", migration.Timestamp))
	}
	return msg.Bytes(), nil
}
