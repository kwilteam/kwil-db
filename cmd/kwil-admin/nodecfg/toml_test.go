package nodecfg

import (
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"

	"github.com/stretchr/testify/assert"
)

func Test_Generate_TOML(t *testing.T) {
	cfg := config.DefaultConfig()

	cfg.AppCfg.DBHost = "/tmp/custom_pg_socket_path"
	cfg.AppCfg.GrpcListenAddress = "localhost:9000"
	cfg.AppCfg.ExtensionEndpoints = []string{"localhost:9001", "localhost:9002"}
	cfg.Logging.OutputPaths = []string{"stdout", "file"}
	err := WriteConfigFile("test.toml", cfg)
	assert.NoError(t, err)
	defer os.Remove("test.toml")

	updatedcfg := config.DefaultConfig()
	tomlCfg, err := config.LoadConfigFile("test.toml")
	assert.NoError(t, err)

	err = updatedcfg.Merge(tomlCfg)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, cfg.AppCfg.DBHost, updatedcfg.AppCfg.DBHost)
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

	err := GenerateTestnetConfig(&genCfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(genCfg.OutputDir)
}

func Test_IncrementingPorts(t *testing.T) {
	type testcase struct {
		input  string
		amount int
		want   string
	}

	testcases := []testcase{
		{
			input:  "localhost:26656",
			amount: 1,
			want:   "localhost:26657",
		},
		{
			input:  "http://localhost:26656",
			amount: 2,
			want:   "http://localhost:26658",
		},
		{
			input:  "https://localhost:26656",
			amount: -2,
			want:   "https://localhost:26654",
		},
		{
			input:  "tcp://0.0.0.0:26656",
			amount: 3,
			want:   "tcp://0.0.0.0:26659",
		},
	}

	for _, tc := range testcases {
		got, err := incrementPort(tc.input, tc.amount)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}
}
