package database

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
	"strings"
)

func callCmd() *cobra.Command {
	var actionName string

	cmd := &cobra.Command{
		Use:   "call",
		Short: "Execute a query",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				dbId, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return fmt.Errorf("target database not properly specified: %w", err)
				}

				lowerName := strings.ToLower(actionName)

				inputs, err := getInputs(args)
				if err != nil {
					return fmt.Errorf("error getting inputs: %w", err)
				}

				res, err := client.CallAction(ctx, dbId, lowerName, inputs)
				if err != nil {
					return fmt.Errorf("error executing action: %w", err)
				}

				results := res.ExportString()
				printTable(results)
				return nil
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")
	cmd.Flags().StringVarP(&actionName, actionNameFlag, "a", "", "the action name (required)")

	cmd.MarkFlagRequired(actionNameFlag)
	return cmd
}
