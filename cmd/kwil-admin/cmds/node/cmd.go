package node

import (
	"github.com/spf13/cobra"
)

const nodeExplain = "The `node` command is used to control a running Kwil node via its authenticated RPC service."

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: nodeExplain,
	Long:  nodeExplain,
}

func NewNodeCmd() *cobra.Command {
	nodeCmd.AddCommand(
		pingCmd(),
		versionCmd(),
		statusCmd(),
		peersCmd(),
		genAuthKeyCmd(),
	)

	return nodeCmd
}
