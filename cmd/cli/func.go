package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Testf(str string) string {
	return str
}

var testCmd = &cobra.Command{
	Use:     "testf",
	Aliases: []string{"test"},
	Short:   "Reverses a string",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		res := Testf(args[0])
		fmt.Println(res)
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
