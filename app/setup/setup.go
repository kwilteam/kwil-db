package setup

import "github.com/spf13/cobra"

func SetupCmd() *cobra.Command {
	const setupLong = `The setup command provides functions for creating and managing node configuration and data, including:
	- performing quick setup of a standalone Kwil node (init) and Kwil test networks (testnet)
	- resetting node state and all data files (reset)`
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "The setup command provides functions for creating and managing node configuration and data.",
		Long:  setupLong,
	}
	setupCmd.AddCommand(ResetCmd(), TestnetCmd())

	return setupCmd
}
