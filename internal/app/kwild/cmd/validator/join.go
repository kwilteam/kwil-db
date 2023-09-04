package validator

import (
	"encoding/base64"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func joinCmd(cfg *config.KwildConfig) *cobra.Command {
	var appGRPCListenAddr string
	// TODO: read the private key from a file
	cmd := &cobra.Command{
		Use:   "join [JoinerPrivateKey]",
		Short: "Request to join the network as a validator",
		Long:  "The Join command is used to request to join the network as a validator. Joiner private key is required.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Joining the network as a validator...")

			ctx := cmd.Context()

			joinerKeyB, err := base64.StdEncoding.DecodeString(args[0])
			if err != nil {
				return err
			}
			joinerKey, err := crypto.Ed25519PrivateKeyFromBytes(joinerKeyB)
			if err != nil {
				return err
			}

			signer := crypto.NewStdEd25519Signer(joinerKey)
			options := []client.ClientOpt{client.WithSigner(signer)}
			clt, err := client.New(appGRPCListenAddr, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ValidatorJoin(ctx)
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
