package config

import (
	"github.com/spf13/viper"
)

type BInfo struct {
	Version string `json:"version"`
}

var BuildInfo BInfo

func InitBuildInfo() error {
	viper.AddConfigPath(".")
	viper.SetConfigName("buildinfo")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		return err // Returning empty config if error occurs
	}

	err = viper.Unmarshal(&BuildInfo)
	return err
}
