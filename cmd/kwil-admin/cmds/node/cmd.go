package node

import (
	"github.com/spf13/cobra"
)

const nodeExplain = "The `node` command is used to get information about a running Kwil node."

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: nodeExplain,
	Long:  nodeExplain,
}

func NewNodeCmd() *cobra.Command {
	nodeCmd.AddCommand(

		versionCmd(),
		statusCmd(),
		peersCmd(),
		genAuthKeyCmd(),
	)

	return nodeCmd
}
