package validator

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/spf13/cobra"
)

func statusCmd(cfg *config.KwildConfig) *cobra.Command {
	// var validatorURL string
	// var privateKey string
	// var ClientChainRPCURL string
	cmd := &cobra.Command{
		Use:   "status [validatorPublicKey]",
		Short: "Get the status of a validatorJoin request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validatorPublicKey = args[0]
			// Send the validator public key to the server to approve the validator
			// through an RPC call
			// ctx := cmd.Context()
			// cfg := &kc.KwildConfig{
			// 	GrpcURL:           validatorURL,
			// 	PrivateKey:        privateKey,
			// 	ClientChainRPCURL: ClientChainRPCURL,
			// }

			// clt, err := kc.NewClient(ctx, cfg)
			// if err != nil {
			// 	return err
			// }

			// status, err := clt.ValidatorJoinStatus(ctx, []byte(args[0]))
			// if err != nil {
			// 	return err
			// }

			// fmt.Printf("Validator Join Status: \n\tapproved: %d\n\trejected: %d\n\trequired: %d\n\tApprovedValidators: %v\n\tRejectedValidators: %v\n\tStatus: %s\n", status.Approved, status.Rejected, status.Pending, status.ApprovedValidators, status.RejectedValidators, status.Status)
			return nil
		},
	}
	// cmd.Flags().StringVarP(&validatorURL, "validatorURL", "v", "", "Validator URL that you want to join")
	// cmd.Flags().StringVarP(&privateKey, "privateKey", "k", "", "The private key of the wallet that will be used for signing")
	// cmd.Flags().StringVarP(&ClientChainRPCURL, "clientChainRPCURL", "c", "", "The client chain RPC URL")
	return cmd
}
