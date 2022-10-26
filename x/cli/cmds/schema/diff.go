package schema

import (
	"github.com/kwilteam/ksl/sqlspec"

	"github.com/spf13/cobra"
)

func createDiffCmd() *cobra.Command {
	var opts struct {
		From string
		To   string
	}

	cmd := &cobra.Command{
		Use:           "diff",
		Short:         "Compute the diff between a source schema and a target schema.",
		Long:          "`kwil schema diff` calculates the diff between a source schema and a target schema. Source and target can be a database URL or the path to an HCL file.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			from, err := loadRealm(cmd.Context(), opts.From, nil)
			if err != nil {
				return err
			}

			to, err := loadRealm(cmd.Context(), opts.To, nil)
			if err != nil {
				return err
			}

			differ := sqlspec.NewDiffer()
			changes, err := differ.RealmDiff(from, to)
			if err != nil {
				return err
			}

			planner := sqlspec.NewPlanner()
			plan, err := planner.PlanChanges(changes)
			if err != nil {
				return err
			}

			if err := planSummary(cmd, plan); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.From, "from", "", "", "The source schema")
	cmd.Flags().StringVarP(&opts.To, "to", "", "", "The target schema")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}
