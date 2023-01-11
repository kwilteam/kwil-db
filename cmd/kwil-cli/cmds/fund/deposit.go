package fund

import (
	"context"
	"errors"
	"fmt"
	"kwil/cmd/kwil-cli/util"
	"kwil/kwil/client"
	"kwil/x/types/contracts/escrow"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func depositCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "deposit",
		Short: "Deposit funds into the funding pool.",
		Long:  `Deposit funds into the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := client.NewClient(cc, viper.GetViper())
				if err != nil {
					return err
				}

				allowance, err := client.Token.Allowance(client.Config.Address, client.Config.PoolAddress)
				if err != nil {
					return err
				}

				// convert arg 0 to big int
				amount, ok := new(big.Int).SetString(args[0], 10)
				if !ok {
					return fmt.Errorf("error converting %s to big int", args[0])
				}

				// check if allowance >= amount
				if allowance.Cmp(amount) < 0 {
					return fmt.Errorf("not enough tokens to deposit %s (allowance %s)", amount.String(), allowance.String())
				}

				balance, err := client.Token.BalanceOf(client.Config.Address)
				if err != nil {
					return err
				}

				if balance.Cmp(amount) < 0 {
					return fmt.Errorf("not enough tokens to deposit %s (balance %s)", amount.String(), balance.String())
				}

				tokenName := client.Token.Symbol()

				fmt.Printf("You will be depositing $%s %s into funding pool %s\n", amount, tokenName, client.Config.PoolAddress)
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

				// now deposit
				// get the validator address
				depoistRes, err := client.Escrow.Deposit(ctx, &escrow.DepositParams{
					Validator: client.Config.ValidatorAddress,
					Amount:    amount,
				})
				if err != nil {
					return err
				}

				fmt.Printf("Deposit transaction sent. Tx hash: %s", depoistRes.TxHash)

				return nil
			})

		},
	}

	return cmd
}
