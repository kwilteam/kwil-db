package util

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

func PrintlnCheckF(format string, args ...any) {
	fmt.Printf("%s %s\n", color.GreenString("✔"), fmt.Sprintf(format, args...))
}

func WriteConfig(values map[string]any) error {
	vip := viper.New()
	vip.SetConfigFile(viper.ConfigFileUsed())

	if err := vip.ReadInConfig(); err != nil {
		return err
	}
	for k, v := range values {
		vip.Set(k, v)
	}
	return vip.WriteConfig()
}
