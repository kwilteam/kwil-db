package validator

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func leaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave [validatorPublicKey]",
		Short: "Remove the node as a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validatorPublicKey = args[0]
			// Send the validator public key to the server to approve the validator
			// through an RPC call
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				rec, err := client.ValidatorLeave(ctx, []byte(args[0]))
				if err != nil {
					return err
				}
				display.PrintTxResponse(rec)
				return nil
			})
		},
	}
	return cmd
}
