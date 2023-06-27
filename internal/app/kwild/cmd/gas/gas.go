package gas

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	kc "github.com/kwilteam/kwil-db/internal/app/kwild/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func enableGasCmd() *cobra.Command {
	var validatorURL string
	var privateKey string
	var ClientChainRPCURL string
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enables gas prices on the transactions",
		Long:  "The enable command is used to enable the gas costs on all the transactions on the validator nodes.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			receipt, err := clt.UpdateGasCosts(ctx, true)
			if err != nil {
				fmt.Println("Error: ", err)
			}
			fmt.Println("Receipt: ", receipt, "Error: ", err)
			if receipt != nil {
				display.PrintTxResponse(receipt)
			}
			return err
		},
	}
	cmd.Flags().StringVarP(&validatorURL, "validatorURL", "v", "", "Validator URL that you want to join")
	cmd.Flags().StringVarP(&privateKey, "privateKey", "k", "", "The private key of the wallet that will be used for signing")
	cmd.Flags().StringVarP(&ClientChainRPCURL, "clientChainRPCURL", "c", "", "The client chain RPC URL")
	return cmd
}

func disableGasCmd() *cobra.Command {
	var validatorURL string
	var privateKey string
	var ClientChainRPCURL string
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disables gas prices on the transactions",
		Long:  "The disable command is used to disable the gas costs on all the transactions on the validator node.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			receipt, err := clt.UpdateGasCosts(ctx, false)
			if err != nil {
				fmt.Println("Error: ", err)
			}

			if receipt != nil {
				display.PrintTxResponse(receipt)
			}
			return err
		},
	}
	cmd.Flags().StringVarP(&validatorURL, "validatorURL", "v", "", "Validator URL that you want to join")
	cmd.Flags().StringVarP(&privateKey, "privateKey", "k", "", "The private key of the wallet that will be used for signing")
	cmd.Flags().StringVarP(&ClientChainRPCURL, "clientChainRPCURL", "c", "", "The client chain RPC URL")
	return cmd
}
