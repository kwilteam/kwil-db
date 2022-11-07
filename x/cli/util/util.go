package util

import (
	"fmt"
	"math/rand"
	"time"

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

func GenerateNonce(l uint8) string {
	var nonce string
	rand.Seed(time.Now().UnixNano())
	for i := uint8(0); i < l; i++ {
		nonce += string(rune(65 + rand.Intn(26)))
	}
	return nonce
}
