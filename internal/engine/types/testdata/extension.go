package testdata

import "github.com/kwilteam/kwil-db/internal/engine/types"

var (
	ExtensionErc20 = &types.Extension{
		Name: "erc20",
		Initialization: []*types.ExtensionConfig{
			{
				Key:   "address",
				Value: "0x1234567890",
			},
		},
		Alias: "token",
	}
)
