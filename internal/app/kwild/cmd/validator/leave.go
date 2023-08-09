package validator

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func leaveCmd() *cobra.Command {
	// var validatorURL string
	// var privateKey string
	// var ClientChainRPCURL string
	cmd := &cobra.Command{
		Use:   "leave [valPrivateKey] [BcRPCURL]",
		Short: "Request to leave the network as a validator",
		Long:  "The leave command is used to request to leave the network as a validator. It requires the Private key of the node attempting to leave as a Validator and the Blockchain RPC URL are required.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg, err := config.LoadKwildConfig()
			if err != nil {
				return err
			}
			options := []client.ClientOpt{client.WithCometBftUrl(args[1])}

			clt, err := client.New(ctx, cfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ValidatorLeave(ctx, args[0], 0)
			if err != nil {
				return err
			}
			fmt.Println("Transaction hash: ", hash)
			return nil
		},
	}
	// cmd.Flags().StringVarP(&validatorURL, "validatorURL", "v", "", "Validator URL that you want to join")
	// cmd.Flags().StringVarP(&privateKey, "privateKey", "k", "", "The private key of the wallet that will be used for signing")
	// cmd.Flags().StringVarP(&ClientChainRPCURL, "clientChainRPCURL", "c", "", "The client chain RPC URL")
	return cmd
}
