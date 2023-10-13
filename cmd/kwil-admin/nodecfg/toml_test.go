package nodecfg

import (
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"

	"github.com/stretchr/testify/assert"
)

func Test_Generate_TOML(t *testing.T) {
	cfg := config.DefaultConfig()

	cfg.AppCfg.SqliteFilePath = "sqlite.db/randomPath"
	cfg.AppCfg.GrpcListenAddress = "localhost:9000"
	cfg.AppCfg.ExtensionEndpoints = []string{"localhost:9001", "localhost:9002"}
	cfg.Logging.OutputPaths = []string{"stdout", "file"}
	writeConfigFile("test.toml", cfg)
	defer os.Remove("test.toml")

	updatedcfg := config.DefaultConfig()
	err := updatedcfg.ParseConfig("test.toml")
	assert.NoError(t, err)
	assert.Equal(t, cfg.AppCfg.SqliteFilePath, updatedcfg.AppCfg.SqliteFilePath)
	assert.Equal(t, cfg.AppCfg.GrpcListenAddress, updatedcfg.AppCfg.GrpcListenAddress)
	assert.Equal(t, cfg.AppCfg.ExtensionEndpoints, updatedcfg.AppCfg.ExtensionEndpoints)
	assert.Equal(t, cfg.Logging.OutputPaths, updatedcfg.Logging.OutputPaths)
}

func Test_GenerateNodeCfg(t *testing.T) {
	genCfg := NodeGenerateConfig{
		// InitialHeight: 0,
		OutputDir:       "test/trybuild/",
		JoinExpiry:      100,
		WithoutGasCosts: true,
		WithoutNonces:   false,
	}

	err := GenerateNodeConfig(&genCfg)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(genCfg.OutputDir)
}

func Test_GenerateTestnetConfig(t *testing.T) {
	genCfg := TestnetGenerateConfig{
		// InitialHeight:           0,
		NValidators:             2,
		NNonValidators:          1,
		OutputDir:               "test/testnet/",
		StartingIPAddress:       "192.168.12.12",
		PopulatePersistentPeers: true,
		P2pPort:                 26656,
	}

	err := GenerateTestnetConfig(&genCfg)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(genCfg.OutputDir)
}
