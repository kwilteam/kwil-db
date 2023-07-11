package validator

import (
	"fmt"
	"strconv"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	kc "github.com/kwilteam/kwil-db/internal/app/kwild/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func joinCmd() *cobra.Command {
	var validatorURL string
	var privateKey string
	var ClientChainRPCURL string
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

			ctx := cmd.Context()
			cfg := &kc.KwildConfig{
				GrpcURL:           validatorURL,
				PrivateKey:        privateKey,
				ClientChainRPCURL: ClientChainRPCURL,
			}

			clt, err := kc.NewClient(ctx, cfg)
			if err != nil {
				return err
			}

			receipt, err := clt.ValidatorJoin(ctx, []byte(args[0]), power)
			if err != nil {
				return err
			}
			display.PrintTxResponse(receipt)
			return nil
		},
	}
	cmd.Flags().StringVarP(&validatorURL, "validatorURL", "v", "", "Validator URL that you want to join")
	cmd.Flags().StringVarP(&privateKey, "privateKey", "k", "", "The private key of the wallet that will be used for signing")
	cmd.Flags().StringVarP(&ClientChainRPCURL, "clientChainRPCURL", "c", "", "The client chain RPC URL")
	return cmd
}
