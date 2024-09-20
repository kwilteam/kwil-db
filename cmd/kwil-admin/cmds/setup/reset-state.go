package setup

import (
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	resetStateLong = "`reset-state`" + ` deletes the data in Postgres.

Unlike the ` + "`" + `reset` + "`" + ` command, which deletes all of the application data in Postgres and all of the block data
in the root directory, ` + "`" + `reset-state` + "`" + ` only deletes the application data in Postgres. This command is useful if
you want to replay all of the blocks without having to re-download them.`

	resetStateExample = `kwil-admin setup reset-state --host localhost --port 5432 --user kwild --password kwild --dbname kwild`
)

func resetStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset-state",
		Short:   "`reset-state` deletes the data in Postgres.",
		Long:    resetStateLong,
		Example: resetStateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			pgConf, err := common.GetPostgresFlags(cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = resetPGState(cmd.Context(), pgConf)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespString("Reset state in Postgres"))
		},
	}

	common.BindPostgresFlags(cmd)

	return cmd
}
