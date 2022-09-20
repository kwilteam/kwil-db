package plan

import (
	"encoding/json"
	"fmt"

	v0 "github.com/kwilteam/kwil-db/internal/api/v0"
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/kwilteam/kwil-db/internal/dbml"
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
			model, err := dbml.ParseFile(opts.File)
			if err != nil {
				return err
			}

			req := &v0.PlanRequest{
				Db: &v0.Database{},
			}
			_ = req

			data, err := json.MarshalIndent(model, "", "    ")
			if err != nil {
				return err
			}

			fmt.Println(string(data))

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
