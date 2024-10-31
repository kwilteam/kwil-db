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
	// Leader's public key
	Leader types.HexBytes `json:"leader"`
	// List of validators (including the leader)
	Validators []types.Validator `json:"validators"`
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

func DefaultConfig() *Config {
	return &Config{
		LogLevel:  log.LevelInfo,
		LogFormat: log.FormatText,
		PeerConfig: PeerConfig{
			IP:       "0.0.0.0",
			Port:     6600,
			Pex:      true,
			BootNode: "/ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv",
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
	//    These names have no underscores or dashes, to form the least common
	//    denominator between all of toml, env, and flags.

	LogLevel  log.Level  `koanf:"log_level" toml:"log_level"`
	LogFormat log.Format `koanf:"log_format" toml:"log_format"`

	PrivateKey types.HexBytes `koanf:"privkey" toml:"privkey"`

	PeerConfig PeerConfig `koanf:"peer" toml:"peer"`
}

type PeerConfig struct {
	IP       string `koanf:"ip" toml:"ip"`
	Port     uint64 `koanf:"port" toml:"port"`
	Pex      bool   `koanf:"pex" toml:"pex" comment:"enable peer exchange"`
	BootNode string `koanf:"bootnode" toml:"bootnode"` // connection string to the seed node for simplicity.
	// TODO: pubkey@ip:port
	// BootNode is presently a libp2p multiaddress like
	//  /ip4/127.0.0.1/tcp/6601/p2p/16Uiu2HAm8iRUsTzYepLP8pdJL3645ACP7VBfZQ7yFbLfdb7WvkL7
	// but it could be different and we construct this format internally

	// ListenAddr string // "127.0.0.1:6600"
	// BootNodes []string // []string{"127.0.0.1:6601"}
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
