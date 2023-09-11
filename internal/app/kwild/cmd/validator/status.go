package validator

import (
	"encoding/hex"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/client"

	"github.com/spf13/cobra"
)

func statusCmd(cfg *config.KwildConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [validatorPublicKey]",
		Short: "Get the status of a validatorJoin request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			options := []client.ClientOpt{}

			clt, err := client.New(cfg.AppCfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			status, err := clt.ValidatorJoinStatus(ctx, []byte(args[0]))
			if err != nil {
				return err
			}

			fmt.Printf("Candidate: %v (want power %d)\n", hex.EncodeToString(status.Candidate), status.Power)
			for i := range status.Board {
				fmt.Printf(" Validator %x, approved = %v\n", status.Board[i], status.Approved[i])
			}
			return nil
		},
	}
	return cmd
}

func validatorSetCmd(cfg *config.KwildConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get the current validator set",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			options := []client.ClientOpt{}

			clt, err := client.New(cfg.AppCfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			vals, err := clt.CurrentValidators(ctx)
			if err != nil {
				return err
			}
			fmt.Println("Current validator set:")
			for i, v := range vals {
				fmt.Printf("% 3d. %v\n", i, v)
			}
			return nil
		},
	}
	return cmd
}
