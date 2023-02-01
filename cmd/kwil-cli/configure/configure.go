package configure

import (
	"kwil/cmd/kwil-cli/common"

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

			runner := &configPrompter{
				Viper: v,
			}

			// defining the prompts to be run

			// endpoint
			runner.AddPrompt(&common.Prompter{
				Label:   "Endpoint",
				Default: v.GetString("endpoint"),
			}, "endpoint")

			// api key
			runner.AddPrompt(&common.Prompter{
				Label:       "API Key",
				Default:     v.GetString("api-key"),
				MaskDefault: true,
				//ShowLast:    4, // took this out because it causes a new line to be printed on each keystroke.
			}, "api-key")

			// chain code
			runner.AddPrompt(&common.Prompter{
				Label:   "Fund Code",
				Default: v.GetString("chain-code"),
			}, "chain-code")

			// eth provider
			runner.AddPrompt(&common.Prompter{
				Label:   "Ethereum Provider Endpoint",
				Default: v.GetString("eth-provider"),
			}, "eth-provider")

			// funding pool address
			runner.AddPrompt(&common.Prompter{
				Label:   "Funding Pool Address",
				Default: v.GetString("funding-pool"),
			}, "funding-pool")

			// private key
			runner.AddPrompt(&common.Prompter{
				Label:       "Private Key",
				Default:     v.GetString("private-key"),
				MaskDefault: true,
			}, "private-key")

			// run the prompts
			if err := runner.Run(); err != nil {
				return err
			}

			return v.WriteConfig()
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
