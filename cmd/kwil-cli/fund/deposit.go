package fund

import (
	"context"
	"errors"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/kwil/client/grpc-client"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func depositCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "deposit",
		Short: "Deposit funds into the funding pool.",
		Long:  `Deposit funds into the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				// @yaiba TODO: no need to dial grpc here, just use the chain client
				client, err := grpc_client.NewClient(cc)
				if err != nil {
					return err
				}

				// convert arg 0 to big int
				amount, ok := new(big.Int).SetString(args[0], 10)
				if !ok {
					return fmt.Errorf("error converting %s to big int", args[0])
				}

				// TODO add tokenName back
				//tokenName := client.Chain.Token.Symbol()

				fmt.Printf("You will be depositing $%s into funding pool %s\n", amount, client.Chain.GetConfig().PoolAddress)
				pr := promptui.Select{
					Label: "Continue?",
					Items: []string{"yes", "no"},
				}

				_, res, err := pr.Run()
				if err != nil {
					return err
				}

				if res != "yes" {
					return errors.New("transaction cancelled")
				}

				txRes, err := client.DepositFund(ctx, client.Chain.GetConfig().PrivateKey, client.Chain.GetConfig().ValidatorAddress, amount)
				if err != nil {
					return err
				}

				fmt.Printf("Deposit transaction sent. Tx hash: %s", txRes.TxHash)
				return nil
			})
		},
	}

	return cmd
}
