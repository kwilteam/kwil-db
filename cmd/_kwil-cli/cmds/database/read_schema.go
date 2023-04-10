package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/engine/types"

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
			ctx := cmd.Context()
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithoutServiceConfig(),
			)
			if err != nil {
				return err
			}

			dbid, err := getSelectedDbid(cmd)
			if err != nil {
				return fmt.Errorf("you must specify either a database name with the --name, or a database id with the --dbid flag")
			}

			meta, err := clt.GetSchema(ctx, dbid)
			if err != nil {
				return err
			}

			// now we print the metadata
			fmt.Println("Tables:")
			for _, t := range meta.Tables {
				fmt.Printf("  %s\n", t.Name)
				fmt.Printf("    Columns:\n")
				for _, c := range t.Columns {
					fmt.Printf("    %s\n", c.Name)
					fmt.Printf("      Type: %s\n", c.Type.String())
					for _, a := range c.Attributes {
						fmt.Printf("      %s\n", a.Type.String())
						value, err := types.NewFromSerial(a.Value)
						if err != nil {
							return err
						}
						if value.IsEmpty() {
							fmt.Printf("        %s\n", value.String())
						}
					}
				}
			}

			// print queries
			fmt.Println("Actions:")
			for _, q := range meta.Actions {
				fmt.Printf("  %s\n", q.Name)
				fmt.Printf("    Type: %s\n", q.Inputs)
			}
			return nil
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "The name of the database to view")
	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database to view(optional, defaults to the your account)")
	cmd.Flags().StringP(dbidFlag, "i", "", "The database id of the database to view(optional, defaults to the your account)")
	return cmd
}
