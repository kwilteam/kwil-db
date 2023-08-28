package validator

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func leaveCmd(cfg *config.KwildConfig) *cobra.Command {
	var appGRPCListenAddr string
	cmd := &cobra.Command{
		Use:   "leave [valPrivateKey] [BcRPCURL]",
		Short: "Request to leave the network as a validator",
		Long:  "The leave command is used to request to leave the network as a validator. It requires the Private key of the node attempting to leave as a Validator and the Blockchain RPC URL are required.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			//options := []client.ClientOpt{client.WithCometBftUrl(args[1])}
			// TODO: Use cometbft client
			options := []client.ClientOpt{}

			clt, err := client.New(appGRPCListenAddr, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ValidatorLeave(ctx, args[0])
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
