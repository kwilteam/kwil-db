package schema

import (
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/kwilteam/kwil-db/internal/schemadef/schema"
	"github.com/kwilteam/kwil-db/internal/sql/postgres"
	"github.com/kwilteam/kwil-db/internal/sql/sqlclient"
	"github.com/spf13/cobra"
)

func createApplyCmd() *cobra.Command {
	var opts struct {
		DatabaseUrl string
		SchemaFiles []string
		AutoApprove bool
	}

	cmd := &cobra.Command{
		Use:           "apply",
		Short:         "Apply a schema to a target database.",
		Long:          "`kwil schema apply' plans and executes a database migration to bring a given database to the state described in the provided schema.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaFile, err := postgres.ParseSchemaFiles(opts.SchemaFiles...)
			if err != nil {
				return err
			}

			client, err := sqlclient.Open(cmd.Context(), opts.DatabaseUrl)
			if err != nil {
				return err
			}
			defer client.Close()

			targetOpts := &schema.InspectRealmOption{}
			if client.URL.Schema != "" {
				targetOpts.Schemas = append(targetOpts.Schemas, client.URL.Schema)
			}
			target, err := client.InspectRealm(cmd.Context(), targetOpts)
			if err != nil {
				return err
			}

			changes, err := client.RealmDiff(target, schemaFile)
			if err != nil {
				return err
			}

			if len(changes) == 0 {
				cmd.Println("Schema is synced, no changes to be made")
				return nil
			}

			plan, err := client.PlanChanges(changes)
			if err != nil {
				return err
			}

			if err := planSummary(cmd, plan); err != nil {
				return err
			}

			if opts.AutoApprove || util.ConfirmPrompt() {
				if err := client.ApplyChanges(cmd.Context(), changes); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&opts.SchemaFiles, "file", "f", nil, "[paths...] file or directory containing the schema definition files")
	cmd.Flags().StringVarP(&opts.DatabaseUrl, "url", "u", "", "URL to the database using the format:\n[driver://username:password@address/dbname?param=value]")
	cmd.Flags().BoolVarP(&opts.AutoApprove, "auto-approve", "y", false, "Auto approve. Apply the schema changes without prompting for approval")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("url")

	return cmd
}
