package nodecfg

import (
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
)

func Test_Generate_TOML(t *testing.T) {
	cfg := config.DefaultConfig()

	cfg.AppCfg.SqliteFilePath = "sqlite.db/randomPath"
	cfg.AppCfg.GrpcListenAddress = "localhost:9000"

	writeConfigFile("test.toml", cfg)

}

func Test_GenerateNodeCfg(t *testing.T) {
	genCfg := NodeGenerateConfig{
		InitialHeight: 0,
		HomeDir:       "test/trybuild/",
	}

	err := GenerateNodeConfig(&genCfg)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(genCfg.HomeDir)
}

func Test_GenerateTestnetConfig(t *testing.T) {
	genCfg := TestnetGenerateConfig{
		NValidators:             2,
		NNonValidators:          1,
		InitialHeight:           0,
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
