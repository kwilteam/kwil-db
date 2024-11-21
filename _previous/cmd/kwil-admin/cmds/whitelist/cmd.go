package whitelist

import "github.com/spf13/cobra"

var peersCmd = &cobra.Command{
	Use:   "whitelist",
	Short: "The whitelist command is used to manage a node's peer whitelist.",
}

func WhitelistCmd() *cobra.Command {
	peersCmd.AddCommand(
		addCmd(),
		removeCmd(),
		listCmd(),
	)
	return peersCmd
}
