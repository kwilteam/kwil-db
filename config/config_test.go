package config

import (
	"fmt"
	"slices"
	"testing"
	"time"

	gotoml "github.com/pelletier/go-toml/v2"

	"github.com/kwilteam/kwil-db/core/log"
)

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
				LogLevel:  log.LevelDebug,
				LogFormat: log.FormatJSON,
				P2P: PeerConfig{
					ListenAddress: "0.0.0.0:6600",
					Pex:           false,
					BootNodes:     []string{"/ip4/192.168.1.1/tcp/8080/p2p/test"},
				},
				DB: DBConfig{
					Host:          "127.0.0.1",
					Port:          "5432",
					User:          "kwild",
					Pass:          "kwild",
					DBName:        "kwild",
					ReadTxTimeout: Duration(45 * time.Second),
					MaxConns:      10,
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
				if loaded.P2P.ListenAddress != tt.config.P2P.ListenAddress {
					t.Errorf("P2P.ListenAddress mismatch: got %v, want %v", loaded.P2P.ListenAddress, tt.config.P2P.ListenAddress)
				}
				if loaded.P2P.Pex != tt.config.P2P.Pex {
					t.Errorf("P2P.Pex mismatch: got %v, want %v", loaded.P2P.Pex, tt.config.P2P.Pex)
				}
				if !slices.Equal(loaded.P2P.BootNodes, tt.config.P2P.BootNodes) {
					t.Errorf("P2P.BootNode mismatch: got %v, want %v", loaded.P2P.BootNodes, tt.config.P2P.BootNodes)
				}
				fmt.Println(loaded.DB)
				if loaded.DB.Host != tt.config.DB.Host {
					t.Errorf("DB.Host mismatch: got %v, want %v", loaded.DB.Host, tt.config.DB.Host)
				}
				if loaded.DB.Port != tt.config.DB.Port {
					t.Errorf("DB.Port mismatch: got %v, want %v", loaded.DB.Port, tt.config.DB.Port)
				}
				if loaded.DB.User != tt.config.DB.User {
					t.Errorf("DB.User mismatch: got %v, want %v", loaded.DB.User, tt.config.DB.User)
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
