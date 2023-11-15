package key

import "github.com/spf13/cobra"

const keyExplain = "The `key` command provides subcommands for private key generation and inspection."

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: keyExplain,
	Long:  "The `key` command provides subcommands for private key generation and inspection. These are the private keys that identify the node on the network and provide validator transaction signing capability.",
}

func NewKeyCmd() *cobra.Command {
	cmd := keyCmd

	// Add subcommands
	cmd.AddCommand(
		genCmd(),
		infoCmd(),
	)

	return cmd
}
