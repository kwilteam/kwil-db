package utils

import (
	"kwil/cmd/kwil-cli/common"

	"github.com/spf13/cobra"
)

func NewCmdUtils() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "utils",
		Aliases: []string{"common"},
		Short:   "Various utility commands.",
		Long:    "",
	}

	cmd.AddCommand(
		signCmd(),
	)

	common.BindKwilFlags(cmd)
	common.BindKwilEnv(cmd)

	return cmd
}
