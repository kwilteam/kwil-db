package role

import (
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/spf13/cobra"
)

func NewCmdRole() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "role",
		Short: "Role is a command that contains subcommands for interacting with roles.",
		Long:  "",
	}

	cmd.AddCommand(
		createRoleCmd(),
		updateRoleCmd(),
		deleteRoleCmd(),
		viewRoleCmd(),
		listRolesCmd(),
	)

	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
