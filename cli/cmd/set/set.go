package set

import (
	"github.com/kwilteam/kwil-db/cli/cmd"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set sets a config value",
	Long: `The "set" command is used to set persistent config values for interacting with Kwil.
With "set", you can set things like the target URL, your ETH provider, private key, etc.`,
}

func init() {
	cmd.RootCmd.AddCommand(setCmd)
}

/*
type promptContent struct {
	errorMsg string
	label    string
}

func promptGetInput(pc promptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Success: "{{ . | bold }}: ",
		Valid:   "{{ . | green }}: ",
		Invalid: "{{ . | red }}: ",
		Prompt:  "{{ . }}: ",
	}

	prompt := promptui.Prompt{
		Label:     pc.label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v", err)
		return ""
	}

	fmt.Printf("Input: %q")
	return result
}
*/
