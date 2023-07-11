package validator

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	kc "github.com/kwilteam/kwil-db/internal/app/kwild/client"
	"github.com/spf13/cobra"
)

func leaveCmd() *cobra.Command {
	var validatorURL string
	var privateKey string
	var ClientChainRPCURL string
	cmd := &cobra.Command{
		Use:   "leave [validatorPublicKey]",
		Short: "Remove the node as a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validatorPublicKey = args[0]
			// Send the validator public key to the server to approve the validator
			// through an RPC call
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

			rec, err := clt.ValidatorLeave(ctx, []byte(args[0]))
			if err != nil {
				return err
			}
			display.PrintTxResponse(rec)
			return nil
		},
	}
	cmd.Flags().StringVarP(&validatorURL, "validatorURL", "v", "", "Validator URL that you want to join")
	cmd.Flags().StringVarP(&privateKey, "privateKey", "k", "", "The private key of the wallet that will be used for signing")
	cmd.Flags().StringVarP(&ClientChainRPCURL, "clientChainRPCURL", "c", "", "The client chain RPC URL")
	return cmd
}
