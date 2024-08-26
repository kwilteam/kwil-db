package peers

import "github.com/spf13/cobra"

var peersCmd = &cobra.Command{
	Use:     "peers",
	Short:   "manages the node's peers",
	Aliases: []string{"peer"},
}

func PeersCmd() *cobra.Command {
	peersCmd.AddCommand(
		addCmd(),
		removeCmd(),
		listCmd(),
	)
	return peersCmd
}
