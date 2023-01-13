package database

import (
	"context"
	"fmt"
	"kwil/kwil/client/grpc-client"

	"kwil/cmd/kwil-cli/common"
	"kwil/x/types/databases"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func viewDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "view",
		Short: "View is used to view the details of a database.  It requires a database owner and a name",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				c, err := grpc_client.NewClient(cc, viper.GetViper())
				if err != nil {
					return fmt.Errorf("error creating client: %w", err)
				}

				meta, err := c.Txs.GetSchema(ctx, &databases.DatabaseIdentifier{
					Owner: args[0],
					Name:  args[1],
				})

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
			})
		},
	}

	return cmd
}

/*
func listDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List is used to list all databases.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				ksc := apipb.NewKwilServiceClient(cc)
				resp, err := ksc.ListDatabases(ctx, &apipb.ListDatabasesRequest{})
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
*/
