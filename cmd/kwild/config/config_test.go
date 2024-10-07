package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/common/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Config_Toml ensures that the various mapstructure tags are set
// correctly, as per the cmd/kwild/config/test_data/config.toml file.
func Test_Config_Toml(t *testing.T) {
	cfg := cmd.DefaultConfig()

	tomlCfg, err := LoadConfigFile(ConfigFilePath("test_data"))
	assert.NoError(t, err)

	err = cfg.Merge(tomlCfg)
	assert.NoError(t, err)

	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "info", cfg.Logging.RPCLevel)
	assert.Equal(t, "warn", cfg.Logging.ConsensusLevel)

	assert.Equal(t, "192.168.1.1:8484", cfg.AppConfig.JSONRPCListenAddress)

	assert.Equal(t, int64(123_000), cfg.Logging.MaxLogSizeKB)

	assert.Equal(t, 6, cfg.Logging.MaxLogRolls)

	// extension endpoints
	assert.Equal(t, 3, len(cfg.AppConfig.ExtensionEndpoints))
	assert.Equal(t, "localhost:50052", cfg.AppConfig.ExtensionEndpoints[0])
	assert.Equal(t, "localhost:50053", cfg.AppConfig.ExtensionEndpoints[1])

	assert.Equal(t, "somewhere.cert", cfg.AppConfig.AdminTLSCertFile)
	assert.Equal(t, "somewhere.key", cfg.AppConfig.AdminTLSKeyFile)

	assert.Equal(t, true, cfg.Instrumentation.Prometheus)
	assert.Equal(t, 6, cfg.Instrumentation.MaxConnections)
	assert.Equal(t, "9.8.7.6:20660", cfg.Instrumentation.PromListenAddr)

	// TODO: Add bunch of other validations for different types
	// 1h -> 3600s
	assert.Equal(t, 1*time.Hour, cfg.ChainConfig.StateSync.TrustPeriod.Dur())
	assert.Equal(t, config.Duration(1*time.Hour), cfg.ChainConfig.StateSync.TrustPeriod)
}

func Test_cleanListenAddr(t *testing.T) {
	type args struct {
		listen        string
		defaultListen string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"orig ok",
			args{
				listen:        "127.0.0.2:9090",
				defaultListen: "127.0.0.1:8080",
			},
			"127.0.0.2:9090",
		},
		{
			"orig lacks port",
			args{
				listen:        "127.0.0.2",
				defaultListen: "127.0.0.1:8080",
			},
			"127.0.0.2:8080",
		},
		{
			"orig lacks IP",
			args{
				listen:        ":9090",
				defaultListen: "127.0.0.1:8080",
			},
			"127.0.0.1:9090",
		},
		{
			"ipv6 too many colons, fallback to input",
			args{
				listen:        "2f:2f::",
				defaultListen: "127.0.0.1:8080",
			},
			"2f:2f::",
		},
		{
			"ipv6 bracketed no port",
			args{
				listen:        "[2f:2f::]",
				defaultListen: "127.0.0.1:8080",
			},
			"[2f:2f::]:8080",
		},
		{
			"ipv6 bracketed with port",
			args{
				listen:        "[2f:2f::]:9090",
				defaultListen: "127.0.0.1:8080",
			},
			"[2f:2f::]:9090",
		},
		{
			"ipv6 default, input with only port",
			args{
				listen:        ":9090",
				defaultListen: "[::1]:8080",
			},
			"[::1]:9090",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanListenAddr(tt.args.listen, tt.args.defaultListen); got != tt.want {
				t.Errorf("cleanListenAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Rootify(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	rootDir := "/app/kwild"

	cwd, err := os.Getwd()
	require.NoError(t, err)

	testcases := []struct {
		name        string
		addr        string
		rootifyPath string
		expandPath  string
	}{
		{
			name:        "absolute path",
			addr:        "/app/kwild",
			rootifyPath: "/app/kwild",
			expandPath:  "/app/kwild",
		},
		{
			name:        "absolute path with tilde",
			addr:        "~/kwild",
			rootifyPath: filepath.Join(homeDir, "kwild"),
			expandPath:  filepath.Join(homeDir, "kwild"),
		},
		{
			name:        "relative path",
			addr:        "genesis.json",
			rootifyPath: "/app/kwild/genesis.json",
			expandPath:  filepath.Join(cwd, "genesis.json"),
		},
		{
			name:        "relative path with ../",
			addr:        "../conf/genesis.json",
			rootifyPath: "/app/conf/genesis.json",
			expandPath:  filepath.Join(cwd, "../conf/genesis.json"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.CleanPath(tc.addr, rootDir)
			require.NoError(t, err)
			assert.Equal(t, tc.rootifyPath, got)

			got, err = ExpandPath(tc.addr)
			require.NoError(t, err)
			assert.Equal(t, tc.expandPath, got)
		})
	}
}
