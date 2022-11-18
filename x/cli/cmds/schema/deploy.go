package schema

import (
	"context"
	"ksl/ast"
	"kwil/x/cli/util"
	"kwil/x/proto/apipb"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func createDeployCmd() *cobra.Command {
	var opts struct {
		DatabaseUrl string
		SchemaFiles []string
		Wallet      string
		Database    string
		AutoApprove bool
	}

	cmd := &cobra.Command{
		Use:           "deploy",
		Short:         "Deploy a schema to a target database.",
		Long:          "`kwil schema deploy' plans and executes a database migration to bring a given database to the state described in the provided schema.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ksch := ast.ParseFiles(opts.SchemaFiles...)
			if ksch.HasErrors() {
				return ksch.Diagnostics
			}
			schemaData := ksch.Data()

			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client := apipb.NewKwilServiceClient(cc)
				req := &apipb.PlanSchemaRequest{Wallet: opts.Wallet, Database: opts.Database, Schema: schemaData}
				planResponse, err := client.PlanSchema(ctx, req)
				if err != nil {
					return err
				}
				planSummaryProto(cmd, planResponse.Plan)
				if opts.AutoApprove || util.ConfirmPrompt() {
					req := &apipb.ApplySchemaRequest{Wallet: opts.Wallet, Database: opts.Database, Schema: schemaData}
					_, err := client.ApplySchema(cmd.Context(), req)
					if err != nil {
						return err
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().StringSliceVarP(&opts.SchemaFiles, "file", "f", nil, "[paths...] file or directory containing the schema definition files")
	cmd.Flags().StringVarP(&opts.DatabaseUrl, "url", "u", "", "URL to the database using the format:\n[driver://username:password@address/dbname?param=value]")
	cmd.Flags().BoolVarP(&opts.AutoApprove, "auto-approve", "y", false, "Auto approve. Apply the schema changes without prompting for approval")
	cmd.Flags().StringVarP(&opts.Wallet, "wallet", "w", "", "Wallet to use for the connection")
	cmd.Flags().StringVarP(&opts.Database, "database", "d", "", "Database name to connect to")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("wallet")
	cmd.MarkFlagRequired("database")
	cmd.MarkFlagRequired("url")

	return cmd
}
