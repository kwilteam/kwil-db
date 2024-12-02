package database

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/client"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return client.DialClient(cmd.Context(), cmd, client.WithoutPrivateKey, func(ctx context.Context, client clientType.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				schema, err := client.GetSchema(ctx, dbid)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting schema: %w", err))
				}

				return display.PrintCmd(cmd, &respSchema{Schema: schema})
			})
		},
	}

	bindFlagsTargetingDatabase(cmd)
	return cmd
}
