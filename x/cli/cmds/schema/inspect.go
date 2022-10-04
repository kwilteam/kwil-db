package schema

import (
	_ "kwil/x/schemadef/postgres"
	"kwil/x/schemadef/schema"
	"kwil/x/sql/sqlclient"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func createInspectCmd() *cobra.Command {
	var opts struct {
		URL     string
		Schemas []string
		Exclude []string
	}

	var cmd = &cobra.Command{
		Use:           "inspect",
		Short:         "Inspect a database and print its schema in Kwil DDL syntax.",
		Long:          "",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := sqlclient.Open(cmd.Context(), opts.URL)
			if err != nil {
				return err
			}
			defer client.Close()
			schemas := opts.Schemas
			if client.URL.Schema != "" {
				schemas = append(schemas, client.URL.Schema)
			}
			s, err := client.InspectRealm(cmd.Context(), &schema.InspectRealmOption{
				Schemas: schemas,
				Exclude: opts.Exclude,
			})
			if err != nil {
				return err
			}
			ddl, err := client.MarshalSpec(s)
			if err != nil {
				return err
			}
			cmd.Print(string(ddl))
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.URL, "url", "u", "", "[driver://username:password@protocol(address)/dbname?param=value] select a database using the URL format")
	cmd.Flags().StringSliceVarP(&opts.Schemas, "schema", "s", nil, "Set schema name")
	cmd.Flags().StringSliceVarP(&opts.Exclude, "exclude", "", nil, "List of glob patterns used to filter resources from inspection")
	cobra.CheckErr(cmd.MarkFlagRequired("url"))
	return cmd
}
