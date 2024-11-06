package node

import (
	"encoding/hex"
	"kwil/log"
	"kwil/node/types"
	"slices"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	gotoml "github.com/pelletier/go-toml/v2"
)

func mustDecodeHex(s string) types.HexBytes {
	dec, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return dec
}

func TestMarshalConfig(t *testing.T) {
	cfg := Config{
		LogLevel:   log.LevelInfo,
		LogFormat:  log.FormatUnstructured,
		PrivateKey: mustDecodeHex("ababababab"),
		PeerConfig: PeerConfig{
			IP:        "127.0.0.1",
			Port:      6600,
			Pex:       true,
			BootNodes: []string{"/ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv"},
		},
	}

	inToml, err := gotoml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("gotoml:\n" + string(inToml))

	k := koanf.New(".")
	err = k.Load(rawbytes.Provider(inToml), toml.Parser(), koanf.WithMergeFunc(func(src, dest map[string]interface{}) error {
		// remove underscores from keys so all sources have the same keys, and
		// we can unmarshal into the Config struct with koanf tags, where there
		// are no underscores.
		for k, v := range src {
			dest[strings.ReplaceAll(k, "_", "")] = v
		}
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	k.Print()

	var inCfg Config
	k.UnmarshalWithConf("", &inCfg, koanf.UnmarshalConf{Tag: "koanf"})

	k.KeyMap()
	spew.Dump(inCfg)

	// k.Marshal just marshals the internal map[string]interface{}:
	// outTomlK, err := k.Marshal(toml.Parser())
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log("ktoml:\n" + string(outTomlK))

	// go-toml.Marshal uses the "toml" tags of the Config struct:
	outToml, err := gotoml.Marshal(inCfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("gotoml:\n" + string(outToml))
}

var testConfigToml = `
log_level = 'info'
log_format = 'plain'
private_key = "ababababab"

[peer]
ip = '127.0.0.1'
port = 6600
pex = true
bootnode = '/ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv'
`

func TestConfigSaveAndLoad(t *testing.T) {
	tempFile := t.TempDir() + "/config.toml"

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				LogLevel:   log.LevelDebug,
				LogFormat:  log.FormatJSON,
				PrivateKey: mustDecodeHex("1234567890"),
				PeerConfig: PeerConfig{
					IP:        "192.168.1.1",
					Port:      8080,
					Pex:       false,
					BootNodes: []string{"/ip4/192.168.1.1/tcp/8080/p2p/test"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty config",
			config: Config{
				LogLevel:  log.LevelInfo,
				LogFormat: log.FormatUnstructured,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.SaveAs(tempFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveAs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			loaded, err := LoadConfig(tempFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if loaded.LogLevel != tt.config.LogLevel {
					t.Errorf("LogLevel mismatch: got %v, want %v", loaded.LogLevel, tt.config.LogLevel)
				}
				if loaded.LogFormat != tt.config.LogFormat {
					t.Errorf("LogFormat mismatch: got %v, want %v", loaded.LogFormat, tt.config.LogFormat)
				}
				if !loaded.PrivateKey.Equals(tt.config.PrivateKey) {
					t.Errorf("PrivateKey mismatch: got %x, want %x", loaded.PrivateKey, tt.config.PrivateKey)
				}
				if loaded.PeerConfig.IP != tt.config.PeerConfig.IP {
					t.Errorf("PeerConfig.IP mismatch: got %v, want %v", loaded.PeerConfig.IP, tt.config.PeerConfig.IP)
				}
				if loaded.PeerConfig.Port != tt.config.PeerConfig.Port {
					t.Errorf("PeerConfig.Port mismatch: got %v, want %v", loaded.PeerConfig.Port, tt.config.PeerConfig.Port)
				}
				if loaded.PeerConfig.Pex != tt.config.PeerConfig.Pex {
					t.Errorf("PeerConfig.Pex mismatch: got %v, want %v", loaded.PeerConfig.Pex, tt.config.PeerConfig.Pex)
				}
				if !slices.Equal(loaded.PeerConfig.BootNodes, tt.config.PeerConfig.BootNodes) {
					t.Errorf("PeerConfig.BootNode mismatch: got %v, want %v", loaded.PeerConfig.BootNodes, tt.config.PeerConfig.BootNodes)
				}
			}
		})
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "non-existent file",
			filename: "nonexistent.toml",
			wantErr:  true,
		},
		{
			name:     "empty filename",
			filename: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfig(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
