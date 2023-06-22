package validator

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [validatorPublicKey]",
		Short: "Get the status of a validatorJoin request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validatorPublicKey = args[0]
			// Send the validator public key to the server to approve the validator
			// through an RPC call
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				status, err := client.ValidatorJoinStatus(ctx, []byte(args[0]))
				if err != nil {
					return err
				}

				fmt.Printf("Validator Join Status: \n\tapproved: %d\n\trejected: %d\n\trequired: %d\n\tApprovedValidators: %v\n\tRejectedValidators: %v\n\tStatus: %s\n", status.Approved, status.Rejected, status.Pending, status.ApprovedValidators, status.RejectedValidators, status.Status)
				return nil
			})
		},
	}
	return cmd
}
