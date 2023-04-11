package config

import (
	"kwil/pkg/config"
)

func LoadKwildConfig() (*KwildConfig, error) {
	cfg := &KwildConfig{}

	err := config.LoadConfig(RegisteredVariables, EnvPrefix, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
