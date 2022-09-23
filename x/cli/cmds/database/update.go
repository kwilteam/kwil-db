package database

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v0 "kwil/x/api/v0"
	"kwil/x/cli/util"
)

func updateDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update is used to modify a database.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// if len(args) != 1 {
			// 	fmt.Println("Please provide a database name")
			// 	return
			// }

			// // TODO: check if database exists

			// input, err := promptModify(args[0])
			// if err != nil {
			// 	fmt.Printf("Prompt failed %v\n", err)
			// 	return
			// }

			// switch input {
			// case "tables":
			// 	table.Table()
			// case "roles":
			// 	roles.Roles()
			// case "queries":
			// 	queries.Queries()
			// }

			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc v0.KwilServiceClient) error {
				resp, err := ksc.UpdateDatabase(ctx, &v0.UpdateDatabaseRequest{})
				if err != nil {
					return err
				}
				_ = resp
				return nil
			})
		},
	}

	return cmd
}
