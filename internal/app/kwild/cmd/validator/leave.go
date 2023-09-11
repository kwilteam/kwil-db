package validator

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"github.com/spf13/cobra"
)

func leaveCmd(cfg *config.KwildConfig) *cobra.Command {
	var appGRPCListenAddr string
	// TODO: read the private key from a file
	cmd := &cobra.Command{
		Use:   "leave [valPrivateKey]",
		Short: "Request to leave the network as a validator",
		Long:  "The leave command is used to request to leave the network as a validator. It requires the Private key of the node attempting to leave.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			leaverKey, err := crypto.Ed25519PrivateKeyFromHex(args[0])
			if err != nil {
				return err
			}
			signer := crypto.NewStdEd25519Signer(leaverKey)
			options := []client.ClientOpt{client.WithSigner(signer)}

			clt, err := client.New(appGRPCListenAddr, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ValidatorLeave(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("Transaction hash: %x\n", hash)
			return nil
		},
	}
	cmd.Flags().StringVar(&appGRPCListenAddr, "grpc_listen_addr", cfg.AppCfg.GrpcListenAddress, "gRPC server address")
	return cmd
}
