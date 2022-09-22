package plan

import (
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/spf13/cobra"
)

func NewCmdPlan() *cobra.Command {
	type Options struct {
		File string
	}

	var opts Options

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan generates a plan for the specified data model",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			// return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, client v0.KwilServiceClient) error {

			// 	return nil
			// })
			return nil
		},
	}

	util.BindKwilFlags(cmd.PersistentFlags())
	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "the datamodel file to plan")
	cmd.MarkFlagRequired("file")

	return cmd
}
