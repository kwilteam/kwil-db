package whitelist

import "github.com/spf13/cobra"

var peersCmd = &cobra.Command{
	Use:   "whitelist",
	Short: "manages the node's whitelist peers",
}

func WhitelistCmd() *cobra.Command {
	peersCmd.AddCommand(
		addCmd(),
		removeCmd(),
		listCmd(),
	)
	return peersCmd
}
