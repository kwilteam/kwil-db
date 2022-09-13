package set

import (
	"fmt"

	cmd "github.com/kwilteam/kwil-db/internal/cli/commands"
	"github.com/kwilteam/kwil-db/internal/cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get gets a config value",
	Long:  `The "get" command is used to get persistent values.`,
}

func init() {
	cmd.RootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getEnvCmd)
}

var getEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Get the current environment variables",
	Long:  `Gets all variables that are currently set.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to make sure there are 0 args
		if len(args) != 0 {
			fmt.Println("get env takes no arguments")
			return
		}

		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		keys := viper.AllKeys()
		for i := 0; i < len(keys); i++ {
			fmt.Printf(`%s="%s"`, keys[i], viper.GetString(keys[i]))
			fmt.Println() // waits and prints a new line
		}
	},
}
