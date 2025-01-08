package rpc

import (
	"github.com/spf13/cobra"
)

const adminExplain = "The `admin` command is used to get information about a running Kwil node."

func NewAdminCmd() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "commands for admin RPCs",
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
	// display.BindOutputFormatFlag(adminCmd)

	return adminCmd
}
