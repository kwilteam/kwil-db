package utils

import (
	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/spf13/cobra"
)

func NewCmdUtils() *cobra.Command {
	var utilsCmd = &cobra.Command{
		Use:     "utils",
		Aliases: []string{"common"},
		Short:   "Various admin utility commands.",
		Long:    "Various admin utility commands.",
	}

	utilsCmd.AddCommand(
		txQueryCmd(),
	)

	rpc.BindRPCFlags(utilsCmd)
	display.BindOutputFormatFlag(utilsCmd)

	return utilsCmd
}
