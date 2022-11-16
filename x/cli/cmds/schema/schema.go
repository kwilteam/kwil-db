package schema

import (
	"github.com/spf13/cobra"
)

func NewCmdSchema() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "schema",
		Short: "Work with Kwil schemas.",
		Long:  "The `kwil schema` command groups subcommands for working with Kwil schemas.",
	}

	cmd.AddCommand(
		createPlanCmd(),
		createApplyCmd(),
	)

	return cmd
}
