package validator

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func approveCmd(cfg *config.KwildConfig) *cobra.Command {
	var appGRPCListenAddr string
	// TODO: read the private key from a file
	cmd := &cobra.Command{
		Use:   "approve [JoinerPublicKey] [ApproverPrivateKey]",
		Short: "Add the validator to the list of approved validators",
		Long:  "The approve command is used to issue a transaction to approve a joining node as a validator. It requires the public key of the joining node and the private key of the approving node. Both keys are base64.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			approverKey, err := crypto.Ed25519PrivateKeyFromHex(args[1])
			if err != nil {
				return err
			}
			signer := crypto.NewStdEd25519Signer(approverKey)
			options := []client.ClientOpt{client.WithSigner(signer)}
			clt, err := client.New(appGRPCListenAddr, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ApproveValidator(ctx, []byte(args[0]))
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
