package config

import (
	"encoding/hex"
	"fmt"
	"kwil/log"
	"kwil/node/types"
	"slices"
	"strings"
	"testing"
	"time"

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

// TestMarshalDuration ensures that a time.Duration can be marshaled and
// unmarshaled with the pelletier/go-toml/v2 library. This wasn't always the
// case for some reason **cough specs cough**.
func TestMarshalDuration(t *testing.T) {
	type td struct {
		Duration Duration `koanf:"duration" toml:"duration"`
	}
	tt := td{
		Duration: Duration(10 * time.Second),
	}
	bts, err := gotoml.Marshal(tt)
	if err != nil {
		t.Fatal(err)
	}

	var tt2 td
	err = gotoml.Unmarshal(bts, &tt2)
	if err != nil {
		t.Fatal(err)
	}
	if tt2.Duration != tt.Duration {
		t.Fatalf("got %v, want %v", tt2.Duration, tt.Duration)
	}
}

func TestMarshalConfig(t *testing.T) {
	cfg := Config{
		LogLevel:   log.LevelInfo,
		LogFormat:  log.FormatUnstructured,
		PrivateKey: mustDecodeHex("ababababab"),
		P2P: PeerConfig{
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

// nolint I'm going to use this
var testConfigToml = `
log_level = 'info'
log_format = 'plain'
private_key = "ababababab"

[pg]
host = '127.0.0.1'
port = 5435
user = 'kwild'
pass = 'kwild'
dbname = 'kwild'
max_connections = 10

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
				P2P: PeerConfig{
					IP:        "192.168.1.1",
					Port:      8080,
					Pex:       false,
					BootNodes: []string{"/ip4/192.168.1.1/tcp/8080/p2p/test"},
				},
				PGConfig: PGConfig{
					Host:           "127.0.0.1",
					Port:           "5432",
					User:           "kwild",
					Pass:           "kwild",
					DBName:         "kwild",
					MaxConnections: 10,
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
				if loaded.P2P.IP != tt.config.P2P.IP {
					t.Errorf("P2P.IP mismatch: got %v, want %v", loaded.P2P.IP, tt.config.P2P.IP)
				}
				if loaded.P2P.Port != tt.config.P2P.Port {
					t.Errorf("P2P.Port mismatch: got %v, want %v", loaded.P2P.Port, tt.config.P2P.Port)
				}
				if loaded.P2P.Pex != tt.config.P2P.Pex {
					t.Errorf("P2P.Pex mismatch: got %v, want %v", loaded.P2P.Pex, tt.config.P2P.Pex)
				}
				if !slices.Equal(loaded.P2P.BootNodes, tt.config.P2P.BootNodes) {
					t.Errorf("P2P.BootNode mismatch: got %v, want %v", loaded.P2P.BootNodes, tt.config.P2P.BootNodes)
				}
				fmt.Println(loaded.PGConfig)
				if loaded.PGConfig.Host != tt.config.PGConfig.Host {
					t.Errorf("PGConfig.Host mismatch: got %v, want %v", loaded.PGConfig.Host, tt.config.PGConfig.Host)
				}
				if loaded.PGConfig.Port != tt.config.PGConfig.Port {
					t.Errorf("PGConfig.Port mismatch: got %v, want %v", loaded.PGConfig.Port, tt.config.PGConfig.Port)
				}
				if loaded.PGConfig.User != tt.config.PGConfig.User {
					t.Errorf("PGConfig.User mismatch: got %v, want %v", loaded.PGConfig.User, tt.config.PGConfig.User)
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
