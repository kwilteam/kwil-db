package schema

import (
	"context"
	"kwil/x/cli/util"
	"kwil/x/proto/schemapb"

	"github.com/kwilteam/ksl/kslparse"
	"github.com/kwilteam/ksl/sqlspec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func createPlanCmd() *cobra.Command {
	var opts struct {
		SchemaFiles []string
		Wallet      string
		Database    string
		AutoApprove bool
	}

	cmd := &cobra.Command{
		Use:           "plan",
		Short:         "Validate a schema and create an execution plan.",
		Long:          "`kwil schema plan' plans the steps required to bring a database to the state described in the provided schema.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fs, err := kslparse.ParseKwilFiles(opts.SchemaFiles...)
			if err != nil {
				return err
			}
			_, diags := sqlspec.Decode(fs)
			if diags.HasErrors() {

				return diags
			}

			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client := schemapb.NewSchemaServiceClient(cc)
				req := &schemapb.PlanSchemaRequest{Wallet: opts.Wallet, Database: opts.Database, Schema: fs.Data()}
				resp, err := client.PlanSchema(ctx, req)
				if err != nil {
					return err
				}
				_ = resp
				return nil
			})
		},
	}

	cmd.Flags().StringSliceVarP(&opts.SchemaFiles, "file", "f", nil, "[paths...] file or directory containing the schema definition files")
	cmd.Flags().BoolVarP(&opts.AutoApprove, "auto-approve", "y", false, "Auto approve. Apply the schema changes without prompting for approval")
	cmd.Flags().StringVarP(&opts.Wallet, "wallet", "w", "", "Wallet to use for the connection")
	cmd.Flags().StringVarP(&opts.Database, "database", "d", "", "Database name to connect to")
	cmd.MarkFlagRequired("file")

	return cmd
}
