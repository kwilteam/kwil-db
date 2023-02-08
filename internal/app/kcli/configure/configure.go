package configure

import (
	"fmt"
	"kwil/internal/app/kcli/common"
	"kwil/internal/app/kcli/config"

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
			fmt.Println("=======", viper.ConfigFileUsed())
			fmt.Println("-------", config.AppConfig.Fund.Wallet)

			runner := &configPrompter{
				Viper: viper.GetViper(),
			}

			// endpoint
			runner.AddPrompt(&common.Prompter{
				Label:   "Endpoint",
				Default: config.AppConfig.Node.Endpoint,
			}, "endpoint")

			// private key
			runner.AddPrompt(&common.Prompter{
				Label:       "Private Key",
				Default:     "sss",
				MaskDefault: true,
			}, "private-key")

			// run the prompts
			if err := runner.Run(); err != nil {
				return err
			}

			return viper.WriteConfig()
		},
	}

	return cmd
}

type configPrompt struct {
	prompt   *common.Prompter
	viperKey string
}

type configPrompter struct {
	prompts []*configPrompt
	Viper   *viper.Viper
}

func (c *configPrompter) AddPrompt(prompt *common.Prompter, viperKey string) {
	c.prompts = append(c.prompts, &configPrompt{
		prompt:   prompt,
		viperKey: viperKey,
	})
}

func (c *configPrompter) Run() error {
	for _, p := range c.prompts {
		res, err := p.prompt.Run()
		if err != nil {
			return err
		}
		c.Viper.Set(p.viperKey, res)
	}
	return nil
}
