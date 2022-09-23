package schema

import (
	"github.com/spf13/cobra"
	"kwil/x/schemadef/postgres"
)

func createDiffCmd() *cobra.Command {
	var opts struct {
		From string
		To   string
	}

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compute the diff between a source schema and a target schema.",
		Long:  "`kwil schema diff` calculates the diff between a source schema and a target schema. Source and target can be a database URL or the path to an HCL file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, err := loadSchema(cmd.Context(), opts.From, nil)
			if err != nil {
				return err
			}

			to, err := loadSchema(cmd.Context(), opts.To, nil)
			if err != nil {
				return err
			}

			differ := postgres.NewDiffer()
			changes, err := differ.SchemaDiff(from, to)
			if err != nil {
				return err
			}

			planner := postgres.NewPlanner()
			plan, err := planner.PlanChanges(cmd.Context(), changes)
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
