package fund

import (
	"errors"
	"fmt"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/kclient"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func depositCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "deposit",
		Short: "Deposit funds into the funding pool.",
		Long:  `Deposit funds into the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// convert arg 0 to big int
			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return fmt.Errorf("error converting %s to big int", args[0])
			}

			clt, err := kclient.New(ctx, config.AppConfig)
			if err != nil {
				return err
			}

			// TODO add tokenName back
			//svcCfg, err := clt.GetServiceConfig(ctx)
			//if err != nil {
			//	return err
			//}
			//tokenName := svcCfg.TokenName

			fmt.Printf("You will be depositing $%s into funding pool %s\n", amount, clt.Config.Fund.PoolAddress)
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

			txRes, err := clt.DepositFund(ctx, amount)
			if err != nil {
				return err
			}

			fmt.Printf("Deposit transaction sent. Tx hash: %s", txRes.TxHash)
			return nil
		},
	}

	return cmd
}
