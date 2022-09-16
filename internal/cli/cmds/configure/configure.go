package configure

import (
	"context"
	"strings"

	"github.com/fatih/color"
	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdConfigure() *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "configure",
		Short:         "Configure your client",
		Long:          "",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.New()
			v.SetConfigFile(viper.ConfigFileUsed())
			if err := v.ReadInConfig(); err != nil {
				return err
			}

			endpointPrompt := util.Prompter{
				Label:   "Endpoint",
				Default: v.GetString("endpoint"),
			}

			apiKeyPrompt := util.Prompter{
				Label:       "API Key",
				Default:     v.GetString("api-key"),
				MaskDefault: true,
				ShowLast:    4,
				HideEntered: true,
			}

			endpoint, err := endpointPrompt.Run()
			if err != nil {
				return err
			}

			apiKey, err := apiKeyPrompt.Run()
			if err != nil {
				return err
			}

			v.Set("endpoint", endpoint)
			v.Set("api-key", apiKey)

			connectPrompt := promptui.Select{
				Label:        "Connect",
				Items:        []string{"yes", "no"},
				HideSelected: true,
			}

			_, doConnect, err := connectPrompt.Run()
			if err != nil {
				return err
			}

			if doConnect == "yes" {
				err := util.ConnectKwil(cmd.Context(), v, func(ctx context.Context, client v0.KwilServiceClient) error {
					res, err := client.Connect(ctx, &v0.ConnectRequest{})
					if err != nil {
						return err
					}
					v.Set("node-address", res.Address)
					util.PrintlnCheckF("Successfully connected to %s", color.YellowString(endpoint))
					util.PrintlnCheckF("Node address is %s", color.YellowString(res.Address))
					return nil
				})

				if err != nil {
					return err
				}

				if err := v.WriteConfig(); err != nil {
					return err
				}
			}

			chainPrompt := promptui.Select{
				Label: "Select Chain",
				Items: []string{"Ethereum", "Goerli"},
			}

			_, chain, err := chainPrompt.Run()
			if err != nil {
				return err
			}

			switch strings.ToLower(chain) {
			case "ethereum":
				v.Set("chain-id", 1)
			case "goerli":
				v.Set("chain-id", 5)
			}

			ethProviderPrompt := util.Prompter{
				Label:   "Ethereum Provider Address",
				Default: v.GetString("eth-provider"),
			}

			ethProvider, err := ethProviderPrompt.Run()
			if err != nil {
				return err
			}

			v.Set("eth-provider", ethProvider)

			fundingPoolAddressPrompt := util.Prompter{
				Label:   "Funding Pool Address",
				Default: v.GetString("funding-pool"),
			}

			fundingPoolAddress, err := fundingPoolAddressPrompt.Run()
			if err != nil {
				return err
			}

			v.Set("funding-pool", fundingPoolAddress)

			privateKeyPrompt := util.Prompter{
				Label:   "Private Key",
				Default: v.GetString("private-key"),
			}

			privateKey, err := privateKeyPrompt.Run()
			if err != nil {
				return err
			}

			v.Set("private-key", privateKey)

			return v.WriteConfig()
		},
	}

	return cmd
}
