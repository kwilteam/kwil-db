package rpc

import (
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/spf13/cobra"
)

const adminExplain = "The `admin` command is used to get information about a running Kwil node using the administration RPC service."

func NewAdminCmd() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "Administrative commands using the secure admin RPC service",
		Long:  adminExplain,
	}

	adminCmd.AddCommand(
		dumpCfgCmd(),
		versionCmd(),
		statusCmd(),
		peersCmd(),
		genAuthKeyCmd(),
	)

	BindRPCFlags(adminCmd)
	display.BindOutputFormatFlag(adminCmd)

	return adminCmd
}
