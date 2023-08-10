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
		GenValidatorCmd(),
		GenNodeKeyCmd(),
		ShowNodeIDCmd(),
		ShowValidatorCmd(),
		NewResetAllCmd(),
		NewResetStateCmd(),
		NewResetPrivValidatorCmd(),
	)

	// NOTE: could add global flags here (home, etc)
	return genCmd
}
