package validator

import (
	"fmt"
	"strconv"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func joinCmd() *cobra.Command {
	// var validatorURL string
	// var privateKey string
	// var ClientChainRPCURL string
	cmd := &cobra.Command{
		Use:   "join [JoinerPrivateKey][power] [BcRPCURL]",
		Short: "Request to join the network as a validator",
		Long:  "The Join command is used to request to join the network as a validator. Joiner Private key and the Blockchain RPC URL are required.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Joining the network as a validator...")
			power, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			fmt.Println("Power:", power, "Pubkey:", args[0])

			ctx := cmd.Context()
			cfg, err := config.LoadKwildConfig()
			if err != nil {
				return err
			}
			options := []client.ClientOpt{client.WithCometBftUrl(args[2])}

			clt, err := client.New(ctx, cfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ValidatorJoin(ctx, args[0], power)
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
