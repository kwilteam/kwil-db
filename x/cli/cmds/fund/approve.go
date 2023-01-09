package fund

import (
	"errors"
	"fmt"
	"kwil/x/cli/client"
	"kwil/x/cli/cmds/display"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func approveCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "approve",
		Short: "Approves the funding pool to spend your tokens",
		Long:  `Approves the funding pool to spend your tokens.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return fmt.Errorf("could not convert %s to int", args[0])
			}

			c, err := client.NewUnconnectedClient(viper.GetViper())
			if err != nil {
				return fmt.Errorf("could not create client: %w", err)
			}

			// get balance
			balance, err := c.Token().BalanceOf(c.Address())
			if err != nil {
				return fmt.Errorf("could not get balance: %w", err)
			}

			// check if balance is less than amount
			if balance.Cmp(amount) < 0 {
				return fmt.Errorf("not enough tokens to fund %s (balance %s)", amount.String(), balance.String())
			}

			cmd.Printf("You will have a new amount approved of: %s\nYour token balance: %s\n", amount.String(), balance.String())

			// ask one more time to confirm the transaction
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

			// approve
			response, err := c.Token().Approve(ctx, c.PoolAddress(), amount)
			if err != nil {
				return err
			}

			display.PrintClientChainResponse(&display.ClientChainResponse{
				Tx:    response.TxHash,
				Chain: c.ChainCode().String(),
			})

			return nil
		},
	}

	return cmd
}
