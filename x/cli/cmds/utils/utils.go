package utils

import (
	"kwil/x/cli/util"

	"github.com/spf13/cobra"
)

func NewCmdUtils() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "utils",
		Aliases: []string{"util"},
		Short:   "Various utility commands.",
		Long:    "",
	}

	cmd.AddCommand(
		signCmd(),
	)

	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
