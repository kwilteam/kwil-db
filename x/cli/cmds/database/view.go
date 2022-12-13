package database

import (
	"context"
	"fmt"

	"kwil/x/cli/util"
	"kwil/x/proto/apipb"

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
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client := apipb.NewKwilServiceClient(cc)
				// should be two args
				if len(args) != 2 {
					return fmt.Errorf("view requires two arguments: owner and name")
				}

				resp, err := client.GetMetadata(ctx, &apipb.GetMetadataRequest{
					Owner:    args[0],
					Database: args[1],
				})

				meta := resp.GetDatabase()

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
						fmt.Printf("      Type: %s\n", c.Type)
						for _, a := range c.Attributes {
							fmt.Printf("      %s\n", a)
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
				for _, q := range meta.Queries {
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
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
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
