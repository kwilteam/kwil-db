package role

import (
	"github.com/spf13/cobra"
)

func updateRoleCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update is used to modify a role.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
