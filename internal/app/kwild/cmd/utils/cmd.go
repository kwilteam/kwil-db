package utils

import (
	"github.com/spf13/cobra"
)

var (
	genCmd = &cobra.Command{
		Use:   "utils",
		Short: "Utility commands to generate and view files required for kwil network",
		Long:  "utils is a command that contains tools to generate and view files required for kwil network",
	}
)

func NewCmdGenerator() *cobra.Command {
	genCmd.AddCommand(
		NewTestnetCmd(),
		InitFilesCmd(),
		ShowNodeIDCmd(),
		NewResetAllCmd(), // TODO: Redo this according to the current files and dir structure
		NewResetStateCmd(),
		NewGenesisHashCmd(),
	)

	// NOTE: could add global flags here (home, etc)
	return genCmd
}
