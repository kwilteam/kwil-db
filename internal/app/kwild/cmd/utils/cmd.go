package utils

import (
	"github.com/spf13/cobra"
)

var (
	genCmd = &cobra.Command{
		Use:   "utils",
		Short: "Utility commands to generate and view files required for kwil network",
		Long:  "utils is a command that contains tools to generate and view files required for kwil network",
		Annotations: map[string]string{
			"skip_load_config": "true",
		},
	}
)

func NewCmdGenerator() *cobra.Command {
	genCmd.AddCommand(
		NewTestnetCmd(),
		InitFilesCmd(),
		GenPrivateKeyCmd(),
		KeyInfoCmd(),
		NewResetAllCmd(), // TODO: Redo this according to the current files and dir structure
		NewResetStateCmd(),
		NewGenesisHashCmd(),
	)

	return genCmd
}
