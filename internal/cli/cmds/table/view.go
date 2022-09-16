package table

import (
	"github.com/spf13/cobra"
)

func viewTableCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "view",
		Short: "View is used to view the details of a table.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

func listTablesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List is used to list the tables in a database.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
