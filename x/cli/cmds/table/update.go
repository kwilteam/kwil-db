package table

import (
	"github.com/spf13/cobra"
)

func updateTableCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update is used to modify a table.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
