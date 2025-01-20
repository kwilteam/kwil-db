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
		Short:   "Miscellaneous utility commands.",
		Long:    "The `utils` commands provide various miscellaneous utility commands such as `query-tx` for querying a transaction status.",
	}

	utilsCmd.AddCommand(
		txQueryCmd(),
	)

	rpc.BindRPCFlags(utilsCmd)
	display.BindOutputFormatFlag(utilsCmd)

	return utilsCmd
}
