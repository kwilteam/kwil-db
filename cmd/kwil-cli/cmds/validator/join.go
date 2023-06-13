package validator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func joinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join [validatorPublicKey] [power]",
		Short: "Request to join the network as a validator",
		Long:  "The Join command is used to request to join the network as a validator. Validator public key and power is required.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Joining the network as a validator...")
			power, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			fmt.Println("Power:", power, "Pubkey:", args[0])
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {

				receipt, err := client.ValidatorJoin(ctx, []byte(args[0]), power)
				if err != nil {
					return err
				}
				display.PrintTxResponse(receipt)
				return nil
			})
		},
	}
	return cmd
}
