package utils

import "github.com/spf13/cobra"

const utilsCmdShort = "The `utils` command is used to get information about a running Kwil node."
const utilsCmdLong = "The `utils` command is used to get information about a running Kwil node."

var utilsCmd = &cobra.Command{
	Use:   "utils",
	Short: utilsCmdShort,
	Long:  utilsCmdLong,
}

func NewUtilsCmd() *cobra.Command {
	utilsCmd.AddCommand(
		pingCmd(),
		queryTxCmd(),
	)

	return utilsCmd
}
