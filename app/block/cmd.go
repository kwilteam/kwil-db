package block

import (
	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "",
	Long:  "",
}

func NewBlockExecCmd() *cobra.Command {
	blockCmd.AddCommand(
		statusCmd(),
		abortCmd(),
	)

	rpc.BindRPCFlags(blockCmd)
	display.BindOutputFormatFlag(blockCmd)
	return blockCmd
}
