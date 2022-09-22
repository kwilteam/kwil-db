package diff

import (
	"github.com/kwilteam/kwil-db/internal/sql/postgres"
	"github.com/spf13/cobra"
)

func NewCmdDiff() *cobra.Command {
	type Options struct {
		Source string
		Target string
	}

	var opts Options

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Calculate and print the diff between two schemas.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := postgres.ParsePaths(opts.Source)
			if err != nil {
				return err
			}

			target, err := postgres.ParsePaths(opts.Target)
			if err != nil {
				return err
			}

			diff := postgres.NewDiff()
			changes, err := diff.RealmDiff(source, target)
			if err != nil {
				return err
			}

		},
	}

	cmd.Flags().StringVarP(&opts.Source, "source", "s", "", "the source schema file")
	cmd.Flags().StringVarP(&opts.Target, "target", "t", "", "the target schema file")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")

	return cmd
}
