package validator

import (
	"fmt"
	"strconv"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func joinCmd(cfg *config.KwildConfig) *cobra.Command {
	var appGRPCListenAddr string
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
			// options := []client.ClientOpt{client.WithCometBftUrl(args[2])}
			// TODO: Use cometbft client
			options := []client.ClientOpt{}
			clt, err := client.New(appGRPCListenAddr, options...)
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
	cmd.Flags().StringVar(&appGRPCListenAddr, "grpc_listen_addr", cfg.AppCfg.GrpcListenAddress, "Address to listen for gRPC connections for the application")
	return cmd
}
