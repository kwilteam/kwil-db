package table

import (
	"github.com/spf13/cobra"
)

func dropTableCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "drop",
		Short: "Drop is used to drop a table.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
