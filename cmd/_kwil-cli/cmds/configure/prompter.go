package configure

import (
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"

	"github.com/spf13/viper"
)

// the configprompter takes a prompt, the viper key that it should set,
// and a list of functions to either validate or modify the input
type configPrompt struct {
	prompt   *common.Prompter
	viperKey string
	fns      []func(res *string) error
}

func (cp *configPrompt) run(v *viper.Viper) error {
	res, err := cp.prompt.Run()
	if err != nil {
		return err
	}

	for _, fn := range cp.fns {
		if err := fn(&res); err != nil {
			fmt.Println(err)
			return cp.run(v)
		}
	}

	v.Set(cp.viperKey, res)
	return nil
}

type configPrompter struct {
	prompts []*configPrompt
	Viper   *viper.Viper
}

// takes a prompt, the viper key to set, and a list of validation functions
func (c *configPrompter) AddPrompt(prompt *common.Prompter, viperKey string, fns ...func(res *string) error) {
	c.prompts = append(c.prompts, &configPrompt{
		prompt:   prompt,
		viperKey: viperKey,
		fns:      fns,
	})
}

func (c *configPrompter) Run() error {
	for _, p := range c.prompts {
		if err := p.run(c.Viper); err != nil {
			return err
		}
	}
	return nil
}
