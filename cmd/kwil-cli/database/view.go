package database

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/cmd/kwil-cli/common"
	"kwil/internal/app/kcli"
)

func viewDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "view",
		Short: "View is used to view the details of a database.  It requires a database owner and a name",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := kcli.New(ctx, common.AppConfig)
			if err != nil {
				return err
			}

			dbName, err := cmd.Flags().GetString("name")
			if err != nil {
				return fmt.Errorf("error getting name flag: %w", err)
			}

			dbOwner, err := cmd.Flags().GetString("owner")
			if err != nil {
				return fmt.Errorf("error getting owner flag: %w", err)
			}
			if dbOwner == "NULL" {
				dbOwner = clt.Config.Fund.GetAccountAddress()
			}

			meta, err := clt.GetDatabaseSchema(ctx, dbOwner, dbName)
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
						if a.Value != nil {
							fmt.Printf("        %s\n", a.Value)
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

	cmd.Flags().StringP("name", "n", "", "The name of the database to view")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("owner", "o", "NULL", "The owner of the database to view")
	return cmd
}
