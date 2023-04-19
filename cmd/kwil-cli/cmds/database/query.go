package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query QUERY_TEXT",
		Short: "Queries a database",
		Long:  "Queries a database. Requires 1 argument: the query text.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return fmt.Errorf("target database not properly specified: %w", err)
				}

				res, err := client.Query(ctx, dbid, args[0])
				if err != nil {
					return fmt.Errorf("error querying database: %w", err)
				}

				counter := 0
				for res.Next() {
					record := res.Record()
					printMap(record.Map())
					counter++
				}
				if counter == 0 {
					fmt.Println("No records found")
				}

				return nil
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")
	return cmd
}

func printMap(m map[string]interface{}) {
	for k, v := range m {
		fmt.Printf("%s: %v\n", k, v)
	}
	fmt.Println()
}
