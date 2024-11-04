package node

import (
	"encoding/json"
	"os"

	"kwil/node/types"
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

//   - toml tags are used to marshal into a toml file with pelletier's go-toml
//     (gotoml.Marshal: Config{} => []byte(tomlString))
//   - koanf tags are used to unmarshal into this struct from a koanf instance
//     (k.Unmarshal: map[string]interface{} => Config{})
//     These names have no underscores or dashes, to form the least common
//     denominator between all of toml, env, and flags.
type NodeConfig struct {
	Port uint64 `json:"port"`
	IP   string `json:"ip"`

	// /ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv
	SeedNode string `json:"seed"` // connection string to the seed node for simplicity.

	Pex bool `json:"pex"` // peer exchange

	PrivateKey types.HexBytes `json:"private_key"`
}

func (nc *NodeConfig) SaveAs(filename string) error {
	bts, err := json.MarshalIndent(nc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, bts, 0644)
}

func LoadNodeConfig(filename string) (*NodeConfig, error) {
	bts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var nc NodeConfig
	if err := json.Unmarshal(bts, &nc); err != nil {
		return nil, err
	}

	return &nc, nil
}
