package role

import (
	"github.com/spf13/cobra"
)

func deleteRoleCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete is used to delete a role.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
