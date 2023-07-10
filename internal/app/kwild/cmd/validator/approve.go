package validator

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

// TODO: List all validators and list all approved nodes
// TODO: If we support revocation, we need to use different way of storing, something like a kv store or something, Also need to remove node from validator set? Only possible in permissioned network

// ApproveCmd is used for approving validators
func approveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [validatorPublicKey]",
		Short: "Add the validator to the list of approved validators",
		Long:  "The approve command is used to approve a validator and add it to the list of the approved validators. Validator public key is required.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validatorPublicKey = args[0]
			// Send the validator public key to the server to approve the validator
			// through an RPC call
			ctx := cmd.Context()
			cfg, err := config.LoadKwildConfig()
			if err != nil {
				return err
			}
			options := []client.ClientOpt{}

			clt, err := client.New(ctx, cfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			err = clt.ApproveValidator(ctx, []byte(args[0]))
			if err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
