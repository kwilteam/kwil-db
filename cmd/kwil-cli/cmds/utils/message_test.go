package utils

import (
	"encoding/hex"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
)

func mustDecodeHex(h string) []byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	return b
}

func Example_respChainInfo_text() {
	display.Print(&respChainInfo{
		&types.ChainInfo{
			ChainID:     "kwil-chain",
			BlockHeight: 100,
			BlockHash:   mustDecodeHex("000000beefbeefbeef"),
		},
	}, nil, "text")
	// Output:
	// Chain ID: kwil-chain
	// Height: 100
	// Hash: 000000beefbeefbeef
}

func Example_respChainInfo_json() {
	display.Print(&respChainInfo{
		&types.ChainInfo{
			ChainID:     "kwil-chain",
			BlockHeight: 100,
			BlockHash:   mustDecodeHex("000000beefbeefbeef"),
		},
	}, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "chain_id": "kwil-chain",
	//     "block_height": 100,
	//     "block_hash": "000000beefbeefbeef"
	//   },
	//   "error": ""
	// }
}
func Example_respKwilCliConfig_text() {
	pk, _, err := crypto.GenerateSecp256k1Key(nil)
	if err != nil {
		panic(err)
	}

	display.Print(&respKwilCliConfig{
		cfg: &config.KwilCliConfig{
			PrivateKey: pk.(*crypto.Secp256k1PrivateKey),
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
	pk, _, err := crypto.GenerateSecp256k1Key(nil)
	if err != nil {
		panic(err)
	}

	display.Print(&respKwilCliConfig{
		cfg: &config.KwilCliConfig{
			PrivateKey: pk.(*crypto.Secp256k1PrivateKey),
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
