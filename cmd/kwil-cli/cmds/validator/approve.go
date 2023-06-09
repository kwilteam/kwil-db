package validator

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func approveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [validatorPublicKey]",
		Short: "Add the validator to the list of approved validators",
		Long:  "The approve command is used to approve a validator and add it to the list of the approved validators. Validator public key is required.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validatorPublicKey = args[0]
			// Send the validator public key to the server to approve the validator
			// through an RPC call
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				err := client.ApproveValidator(ctx, []byte(args[0]))
				if err != nil {
					return err
				}
				return nil
			})
		},
	}
	return cmd
}
