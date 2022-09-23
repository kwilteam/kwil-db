package role

import (
	"github.com/spf13/cobra"
)

func createRoleCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create is used for creating a new role.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
