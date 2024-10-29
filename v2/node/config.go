package node

import (
	"encoding/json"
	"os"
	"p2p/node/types"
)

type GenesisConfig struct {
	// Leader's public key
	Leader []byte `json:"leader"`
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
		return nil, err
	}

	var nc GenesisConfig
	if err := json.Unmarshal(bts, &nc); err != nil {
		return nil, err
	}

	return &nc, nil
}

type NodeConfig struct {
	Port uint64 `json:"port"`
	Ip   string `json:"ip"`

	// /ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv
	SeedNode string `json:"seed"` // connection string to the seed node for simplicity.

	Pex bool `json:"pex"` // peer exchange

	PrivateKey []byte `json:"private_key"`
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
