package nodecfg_test

import (
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/pkg/nodecfg"
)

func Test_Generate_TOML(t *testing.T) {
	cfg := config.DefaultConfig()

	cfg.AppCfg.SqliteFilePath = "sqlite.db/randomPath"
	cfg.AppCfg.GrpcListenAddress = "192.168.5.6"

	nodecfg.WriteConfigFile("test.toml", cfg)

}

func Test_GenerateNodeCfg(t *testing.T) {
	genCfg := nodecfg.NodeGenerateConfig{
		InitialHeight: 0,
		HomeDir:       "test/trybuild/",
	}

	err := nodecfg.GenerateNodeConfig(&genCfg)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(genCfg.HomeDir)
}

func Test_GenerateTestnet(t *testing.T) {
	genCfg := nodecfg.TestnetGenerateConfig{
		NValidators:             2,
		NNonValidators:          1,
		InitialHeight:           0,
		OutputDir:               "test/testnet/",
		StartingIPAddress:       "192.168.1.4",
		PopulatePersistentPeers: true,
		P2pPort:                 26656,
	}

	err := nodecfg.GenerateTestnetConfig(&genCfg)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(genCfg.OutputDir)
}
