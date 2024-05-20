package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Config_Toml(t *testing.T) {
	cfg := DefaultConfig()

	tomlCfg, err := LoadConfigFile(filepath.Join("test_data", ConfigFileName))
	assert.NoError(t, err)

	err = cfg.Merge(tomlCfg)
	assert.NoError(t, err)

	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "info", cfg.Logging.RPCLevel)
	assert.Equal(t, "warn", cfg.Logging.ConsensusLevel)

	assert.Equal(t, "192.168.1.1:8484", cfg.AppCfg.JSONRPCListenAddress)
	assert.Equal(t, "localhost:8080", cfg.AppCfg.HTTPListenAddress)

	// extension endpoints
	assert.Equal(t, 3, len(cfg.AppCfg.ExtensionEndpoints))
	assert.Equal(t, "localhost:50052", cfg.AppCfg.ExtensionEndpoints[0])
	assert.Equal(t, "localhost:50053", cfg.AppCfg.ExtensionEndpoints[1])

	// TODO: Add bunch of other validations for different types
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
