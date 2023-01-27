package chainsync

import (
	chains "kwil/x/chain/types"
)

func (c *chain) ChainCode() chains.ChainCode {
	return c.chainClient.ChainCode()
}
