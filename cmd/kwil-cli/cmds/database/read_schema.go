package database

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

// TODO: @brennan: make the way this prints out the metadata more readable
var (
	readSchemaLong = `Read schema is used to view the details of a deployed database schema.

	You can either specify the database to execute this against with the ` + "`" + `--name` + "`" + ` and ` + "`" + `--owner` + "`" + ` 
	flags, or you can specify the database by passing the database id with the ` + "`" + `--dbid` + "`" + ` flag.  If a ` + "`" + `--name` + "`" + `
	flag is passed and no ` + "`" + `--owner` + "`" + ` flag is passed, the owner will be inferred from your configured wallet.`

	readSchemaExample = `# Reading the schema of the "mydb" database, owned by 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64
kwil-cli database read-schema --name mydb --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64`
)

func readSchemaCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "read-schema",
		Short:   "Read schema is used to view the details of a deployed database schema.",
		Long:    readSchemaLong,
		Example: readSchemaExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client common.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return fmt.Errorf("you must specify either a database name with the --name, or a database id with the --dbid flag")
				}

				schema, err := client.GetSchema(ctx, dbid)
				if err != nil {
					return fmt.Errorf("error getting schema: %w", err)
				}

				return display.PrintCmd(cmd, &respSchema{Schema: schema})
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the target database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the target database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the target database id")
	return cmd
}
