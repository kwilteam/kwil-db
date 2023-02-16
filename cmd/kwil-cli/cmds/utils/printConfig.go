package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printConfigCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "print-config",
		Short: "Print the current configuration",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			printMapRecursively(viper.AllSettings(), 0)
			return nil
		},
	}

	return cmd
}

func printMapRecursively(m map[string]interface{}, indent int) {
	for k, v := range m {
		fmt.Print(strings.Repeat(" ", indent))
		fmt.Printf("%s: ", k)

		switch val := v.(type) {
		case string:
			fmt.Println(val)
		case int:
			fmt.Println(val)
		case map[string]interface{}:
			fmt.Println()
			printMapRecursively(val, indent+2)
		default:
			fmt.Println(val)
		}
	}
}
