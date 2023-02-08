package chainsync

import (
	chains "kwil/pkg/chain/types"
)

func (c *chain) ChainCode() chains.ChainCode {
	return c.chainClient.ChainCode()
}
