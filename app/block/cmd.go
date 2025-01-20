package block

import (
	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Leader block execution commands",
	Long:  "The `block` command group has subcommands for managing leader block execution, including status and aborting.",
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
