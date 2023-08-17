package utils

import (
	"fmt"
	"reflect"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"

	"github.com/spf13/cobra"
)

func printConfigCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "print-config",
		Short: "Print the current configuration",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCliConfig()
			if err != nil {
				return err
			}

			printStruct(cfg.ToPeristedConfig())
			return nil
		},
	}

	return cmd
}

func printStruct(s interface{}) {
	v := reflect.ValueOf(s)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		fmt.Printf("%s: %v\n", field.Name, fieldValue)
	}
}
