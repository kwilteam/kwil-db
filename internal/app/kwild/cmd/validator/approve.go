package validator

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"

	"github.com/spf13/cobra"
)

// TODO: List all validators and list all approved nodes
// TODO: If we support revocation, we need to use different way of storing, something like a kv store or something, Also need to remove node from validator set? Only possible in permissioned network

// ApproveCmd is used for approving validators
func approveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [JoinerPublicKey] [ApproverPrivateKey] [BcRPCURL]",
		Short: "Add the validator to the list of approved validators",
		Long:  "The approve command is used by the Validator node to issue a Approve Transaction to approve a joining node as a validator. It requires the public key of the joining node, the private key of the approving node and the blockchain RPC URL. Both keys are base64.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			/*
				1. Get the public key of the joining node
				2. Get the private key of the approving node
				3. Client that connects to the blockchain rpc interface (cfg)
				4. Send the transaction to the blockchain interface
			*/

			ctx := cmd.Context()
			cfg, err := config.LoadKwildConfig()
			if err != nil {
				return err
			}
			fmt.Println("Arg1: ", args[0], "Arg2: ", args[1])
			fmt.Println(cfg, "BcRPCURL: ", cfg.BcRpcUrl, "GrpcListenAddress: ", cfg.GrpcListenAddress)
			//options := []client.ClientOpt{client.WithCometBftUrl(args[2])}
			options := []client.ClientOpt{}

			clt, err := client.New(cfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			hash, err := clt.ApproveValidator(ctx, args[1], args[0])
			if err != nil {
				return err
			}
			fmt.Println("Transaction hash: ", hash)
			return nil
		},
	}
	return cmd
}
