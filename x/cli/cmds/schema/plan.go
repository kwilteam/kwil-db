package schema

import (
	"github.com/spf13/cobra"
)

func createPlanCmd() *cobra.Command {
	var opts struct {
		SchemaFiles []string
		AutoApprove bool
	}

	cmd := &cobra.Command{
		Use:           "plan",
		Short:         "Validate a schema and create an execution plan.",
		Long:          "`kwil schema plan' plans the steps required to bring a database to the state described in the provided schema.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
			// fs, err := kslparse.ParseKwilFiles(opts.SchemaFiles...)
			// if err != nil {
			// 	return err
			// }

			// data, err := io.ReadAll(rd)
			// if err != nil {
			// 	return err
			// }

			// return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc apipb.KwilServiceClient) error {
			// 	req := &apipb.PlanSchemaRequest{Schema: data}
			// 	resp, err := ksc.PlanSchema(ctx, req)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	_ = resp
			// 	return nil
			// })
		},
	}

	cmd.Flags().StringSliceVarP(&opts.SchemaFiles, "file", "f", nil, "[paths...] file or directory containing the schema definition files")
	cmd.Flags().BoolVarP(&opts.AutoApprove, "auto-approve", "y", false, "Auto approve. Apply the schema changes without prompting for approval")
	cmd.MarkFlagRequired("file")

	return cmd
}
