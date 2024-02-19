package testdata

import types "github.com/kwilteam/kwil-db/common"

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
