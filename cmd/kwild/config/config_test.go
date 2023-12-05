package config_test

import (
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"

	"github.com/stretchr/testify/assert"
)

func Test_Config_Toml(t *testing.T) {
	cfg := config.DefaultConfig()

	tomlCfg, err := config.LoadConfigFile(filepath.Join("test_data", config.ConfigFileName))
	assert.NoError(t, err)

	err = cfg.Merge(tomlCfg)
	assert.NoError(t, err)

	assert.Equal(t, "localhost:50051", cfg.AppCfg.GrpcListenAddress)
	assert.Equal(t, "localhost:8080", cfg.AppCfg.HTTPListenAddress)

	// extension endpoints
	assert.Equal(t, 3, len(cfg.AppCfg.ExtensionEndpoints))
	assert.Equal(t, "localhost:50052", cfg.AppCfg.ExtensionEndpoints[0])
	assert.Equal(t, "localhost:50053", cfg.AppCfg.ExtensionEndpoints[1])

	// TODO: Add bunch of other validations for different types
}
