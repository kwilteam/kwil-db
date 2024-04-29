package utils

import (
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
)

func Example_respChainInfo_text() {
	display.Print(&respChainInfo{
		&types.ChainInfo{
			ChainID:     "kwil-chain",
			BlockHeight: 100,
			BlockHash:   "00000beefbeefbeef",
		},
	}, nil, "text")
	// Output:
	// Chain ID: kwil-chain
	// Height: 100
	// Hash: 00000beefbeefbeef
}

func Example_respChainInfo_json() {
	display.Print(&respChainInfo{
		&types.ChainInfo{
			ChainID:     "kwil-chain",
			BlockHeight: 100,
			BlockHash:   "00000beefbeefbeef",
		},
	}, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "chain_id": "kwil-chain",
	//     "block_height": 100,
	//     "block_hash": "00000beefbeefbeef"
	//   },
	//   "error": ""
	// }
}
func Example_respKwilCliConfig_text() {
	pk, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		panic(err)
	}

	display.Print(&respKwilCliConfig{
		cfg: &config.KwilCliConfig{
			PrivateKey: pk,
			ChainID:    "chainid123",
			Provider:   "localhost:9090",
		},
	}, nil, "text")
	// Output:
	// PrivateKey: ***
	// Provider: localhost:9090
	// ChainID: chainid123
}

func Example_respKwilCliConfig_json() {
	pk, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		panic(err)
	}

	display.Print(&respKwilCliConfig{
		cfg: &config.KwilCliConfig{
			PrivateKey: pk,
			ChainID:    "chainid123",
			Provider:   "localhost:9090",
		},
	}, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "private_key": "***",
	//     "provider": "localhost:9090",
	//     "chain_id": "chainid123"
	//   },
	//   "error": ""
	// }
}
