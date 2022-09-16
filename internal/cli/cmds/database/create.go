package database

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func createDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create is used for creating a new database.",
		Long: `Create is used for creating a new database that will be stored
	under your account.  It takes in one argument, which is the name of the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namePrompt := util.Prompter{Label: "Name"}
			bucketPrompt := util.Prompter{Label: "Bucket Name"}

			dbName, err := namePrompt.Run()
			if err != nil {
				return err
			}

			bucketName, err := bucketPrompt.Run()
			if err != nil {
				return err
			}

			cmd.Println(dbName)
			cmd.Println(bucketName)

			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc v0.KwilServiceClient) error {
				resp, err := ksc.CreateDatabase(ctx, &v0.CreateDatabaseRequest{
					Name: dbName,
				})
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
