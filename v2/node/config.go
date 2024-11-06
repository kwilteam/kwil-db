package node

import (
	"encoding/json"
	"fmt"
	"os"

	"kwil/log"
	"kwil/node/types"

	"github.com/pelletier/go-toml/v2"
)

type GenesisConfig struct {
	// Leader is the leader's public key.
	Leader types.HexBytes `json:"leader"`
	// Validators is the list of genesis validators (including the leader).
	Validators []types.Validator `json:"validators"`

	// TODO: more params like max block size, etc.
}

func (nc *GenesisConfig) SaveAs(filename string) error {
	bts, err := json.MarshalIndent(nc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, bts, 0644)
}

func LoadGenesisConfig(filename string) (*GenesisConfig, error) {
	bts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err // can be os.ErrNotExist
	}

	var nc GenesisConfig
	if err := json.Unmarshal(bts, &nc); err != nil {
		return nil, err
	}

	return &nc, nil
}

// DefaultConfig generates an instance of the default config.
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  log.LevelInfo,
		LogFormat: log.FormatUnstructured,
		// Private key is empty by default.
		PeerConfig: PeerConfig{
			IP:        "0.0.0.0",
			Port:      6600,
			Pex:       true,
			BootNodes: []string{},
		},
	}
}

// Config is the node's config.
type Config struct {
	// NOTE about tags:
	//
	//  - toml tags are used to marshal into a toml file with pelletier's go-toml
	//    (gotoml.Marshal: Config{} => []byte(tomlString))
	//
	//  - koanf tags are used to unmarshal into this struct from a koanf instance
	//    (k.Unmarshal: map[string]interface{} => Config{})
	//
	// Presently these tags are the same. If we change the canonicalization,
	// such as removing both dashes and underscores, the tags would be different.

	LogLevel  log.Level  `koanf:"log_level" toml:"log_level"`
	LogFormat log.Format `koanf:"log_format" toml:"log_format"`

	PrivateKey types.HexBytes `koanf:"privkey" toml:"privkey"`

	PeerConfig PeerConfig `koanf:"peer" toml:"peer"`
}

// PeerConfig corresponds to the [peer] section of the config.
type PeerConfig struct {
	IP        string   `koanf:"ip" toml:"ip"`
	Port      uint64   `koanf:"port" toml:"port"`
	Pex       bool     `koanf:"pex" toml:"pex" comment:"enable peer exchange"`
	BootNodes []string `koanf:"bootnodes" toml:"bootnodes" comment:"list of bootnodes"`
	// TODO: pubkey@ip:port
	// BootNode is presently a libp2p multiaddress like
	//  /ip4/127.0.0.1/tcp/6601/p2p/16Uiu2HAm8iRUsTzYepLP8pdJL3645ACP7VBfZQ7yFbLfdb7WvkL7
	// but it could be different and we construct this format internally

	// ListenAddr string // "127.0.0.1:6600"
}

func (nc *Config) SaveAs(filename string) error {
	bts, err := toml.Marshal(nc)
	if err != nil {
		return err
	}

	// TODO: write a toml header/comment or perhaps use a text/template toml file

	return os.WriteFile(filename, bts, 0644)
}

func LoadConfig(filename string) (*Config, error) {
	bts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(bts))

	var nc Config
	if err := toml.Unmarshal(bts, &nc); err != nil {
		return nil, err
	}

	return &nc, nil
}
