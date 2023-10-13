package database

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/spf13/cobra"
)

// TODO: @brennan: make the way this prints out the metadata more readable
func readSchemaCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "read-schema",
		Short: "Read schema is used to view the details of a database.  It requires a database name",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp respSchema
			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return fmt.Errorf("you must specify either a database name with the --name, or a database id with the --dbid flag")
				}

				resp.Schema, err = client.GetSchema(ctx, dbid)
				if err != nil {
					return fmt.Errorf("error getting schema: %w", err)
				}
				return err
			})

			return display.Print(&resp, err, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "The name of the database to view")
	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database to view(optional, defaults to the your account)")
	cmd.Flags().StringP(dbidFlag, "i", "", "The database id of the database to view(optional, defaults to the your account)")
	return cmd
}
