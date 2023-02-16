package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

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

			meta, err := clt.GetSchemaById(ctx, dbid)
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
						if !a.Value.IsEmpty() {
							fmt.Printf("        %s\n", a.Value.String())
						}
					}
				}
			}

			// print the roles
			fmt.Println("Roles:")
			for _, r := range meta.Roles {
				fmt.Printf("  %s\n", r.Name)
				fmt.Printf("    Permissions:\n")
				for _, p := range r.Permissions {
					fmt.Printf("      %s\n", p)
				}
			}

			// print queries
			fmt.Println("Queries:")
			for _, q := range meta.SQLQueries {
				fmt.Printf("  %s\n", q.Name)
			}

			// Print indexes
			fmt.Println("Indexes:")
			for _, i := range meta.Indexes {
				fmt.Printf("  %s:\n", i.Name)
				fmt.Println("    Type: ", i.Using)
				fmt.Printf("    Table: %s\n", i.Table)
				fmt.Printf("    Columns:\n")
				for _, c := range i.Columns {
					fmt.Printf("      %s\n", c)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "The name of the database to view")
	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database to view(optional, defaults to the your account)")
	cmd.Flags().StringP(dbidFlag, "i", "", "The database id of the database to view(optional, defaults to the your account)")
	return cmd
}
